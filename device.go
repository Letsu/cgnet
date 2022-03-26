package goCisco

import (
	"bufio"
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
	Ip string
	Port string
	Username string
	Password string
	Enable string
	DeviceType string
	ConnType string
	conn net.Conn
	stdin io.Writer
	stdout io.Reader
	readChan chan *string
	prompt string
}

var (
	ErrUnknownCommand = errors.New("unknown or invalid command")
	RegPrompt = regexp.MustCompile("[A-Za-z0-9-_()]+\\#")
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
		break
	default:
		return errors.New("undefined connection type")
	}


	return nil
}

func (d *Device) connectTelnet() error {
	var err error
	if d.Port == "" {
		d.Port = "23"
	}
	d.conn, err = net.Dial("tcp", d.Ip + ":" + d.Port)
	if err != nil {
		return err
	}
	d.stdout = bufio.NewReader(d.conn)
	d.stdin = bufio.NewWriter(d.conn)
	d.readChan = make(chan *string, 20)

	buf := make([]byte, 10000)
	start := false
	for start {
		n, _ := d.stdout.Read(buf)
		start, _ = regexp.MatchString(RegPrompt.String(), string(buf[:n]))
	}

	d.prompt, err = d.Exec("")
	if err != nil {
		log.Println(err)
		return err
	}
	d.Exec("terminal length 0")

	return nil
}

func (d *Device) Exec2(cmd ...string) error {
	_, err := d.Exec(cmd...)
	if err != nil {
		return err
	}
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
			outputFormat = strings.Replace(outputFormat, strings.Join(cmd, ""), "", -1)
			outputFormat = NLStart.ReplaceAllString(outputFormat, "")
			prompt := regexp.MustCompile("[A-Za-z0-9-_()]+\\#")
			outputFormat = prompt.ReplaceAllString(outputFormat, "")
			outputFormat = NLEnd.ReplaceAllString(outputFormat, "")

			if strings.Contains(outputFormat, "Unknown command") || strings.Contains(outputFormat, "Invalid input") {
				return "", ErrUnknownCommand
			}

			return outputFormat, nil
		case <-time.After(time.Second * time.Duration(30)):
			d.Close()
			return "", fmt.Errorf("timeout on %s", d.Ip)
		}
	}
}

func (d *Device) reader(cmd ...string) {
	buf := make([]byte, 10000)
	output := ""

	prompt := regexp.MustCompile("[A-Za-z0-9-_()]+\\#")

	for {
		n, _ := d.stdout.Read(buf)

		output += string(buf[:n])
		if prompt.MatchString(output) && strings.Contains(output, strings.Join(cmd, "")) {
			break
		}
	}

	d.readChan <- &output
}

func (d Device) Close() {
	d.conn.Close()
}