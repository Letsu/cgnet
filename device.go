package goCisco

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type Device struct {
	Ip         string
	Port       string
	Username   string
	Password   string
	Enable     string
	DeviceType string
	ConnType   string
	conn       net.Conn
	stdin      io.Writer
	stdout     io.Reader
	readChan   chan *string
	prompt     string
}

var (
	ErrUnknownCommand = errors.New("unknown or invalid command")
)

func (d *Device) Open() error {
	switch d.ConnType {
	case "telnet":
		err := d.connectTelnet()
		if err != nil {
			return err
		}
		break
	case "ssh":
		err := d.connectSSH()
		if err != nil {
			return err
		}
		break
	default:
		return errors.New("undefined connection type")
	}

	return nil
}

func (d *Device) getPrompt() *regexp.Regexp {
	return regexp.MustCompile(d.prompt + "[[:alnum:]]*[\\#>]")
}

func (d *Device) Exec2(cmd ...string) error {
	_, err := d.Exec(cmd...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Device) login() error {
	buf := make([]byte, 1000)
	n, _ := d.stdout.Read(buf)
	text := string(buf[:n])

	var match bool
	//Login @todo add timeout
	for !match {

		switch {
		case strings.Contains(text, "sername:"):
			io.WriteString(d.conn, d.Username+"\n")
			break
		case strings.Contains(text, "assword:"):
			io.WriteString(d.conn, d.Password+"\n")
			break
		case strings.Contains(text, "timeout"):
			return errors.New("timeout")
		case strings.Contains(text, "Authentication failed"):
			return errors.New("authentication failed")
		default:
			break
		}
		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
		match, _ = regexp.MatchString(d.getPrompt().String(), text)
	}
	d.prompt = d.getPrompt().FindString(text)

	// Enable @todo add timeout
	enabled := !strings.Contains(d.prompt, ">")
	if d.Enable == "" {
		enabled = true
	}

	if !enabled {
		io.WriteString(d.conn, "enable\n")
		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
	}
	for !enabled {
		switch {
		case strings.Contains(text, "assword:"):
			io.WriteString(d.conn, d.Enable+"\n")
			break
		default:
			break
		}

		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
		enabled, _ = regexp.MatchString("[[:alnum:]]*[\\#]", text)
	}

	d.prompt = strings.Replace(d.prompt, ">", "", -1)
	d.prompt = strings.Replace(d.prompt, "#", "", -1)

	return nil
}

func (d *Device) Exec(cmd ...string) (string, error) {
	go d.reader(cmd...)
	_, err := io.WriteString(d.conn, fmt.Sprint(strings.Join(cmd, ""), "\n"))
	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case output := <-d.readChan:
			if output == nil {
				continue
			}

			NLStart := regexp.MustCompile(`^\r?\n`)
			NLEnd := regexp.MustCompile(`\r?\n$`)
			outputFormat := NLStart.ReplaceAllString(*output, "")
			outputFormat = NLStart.ReplaceAllString(outputFormat, "")
			outputFormat = d.getPrompt().ReplaceAllString(outputFormat, "")
			outputFormat = NLEnd.ReplaceAllString(outputFormat, "")

			if strings.Contains(outputFormat, "Unknown command") || strings.Contains(outputFormat, "Invalid input") {
				return "", ErrUnknownCommand
			}

			return outputFormat, nil
		case <-time.After(time.Second * time.Duration(5)):
			return "", fmt.Errorf("no return of prompt on command")
		}
	}
}

func (d *Device) reader(cmd ...string) {
	buf := make([]byte, 10000)
	output := ""

	for {
		n, _ := d.stdout.Read(buf)

		output += string(buf[:n])
		if d.getPrompt().MatchString(output) && strings.Contains(output, strings.Join(cmd, "")) {
			output = strings.Replace(output, strings.Join(cmd, ""), "", -1)
			break
		}
	}

	d.readChan <- &output
}

func (d Device) Close() error {
	close(d.readChan)
	return d.conn.Close()
}
