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
	ErrUnknownCommand = errors.New("unkown command")
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
	d.stdout.Read(buf)


	d.prompt, err = d.Exec("")
	if err != nil {
		log.Println(err)
	}
	_, err = d.Exec("terminal length 0")

	intBrief, err := d.Exec("sh ip int brief")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(intBrief)

	return nil
}

func (d *Device) Exec(cmd ...string) (string, error) {
	go d.reader()
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
			outputFormat = strings.Replace(outputFormat, d.prompt, "", -1)
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

func (d *Device) reader() {
	buf := make([]byte, 10000)
	output := ""

	prompt := regexp.MustCompile("[A-Za-z0-9-_]+\\#")

	for {
		n, err := d.stdout.Read(buf)
		if err != nil {
			log.Println(err)
		}

		output += string(buf[:n])
		if prompt.MatchString(output) {
			break
		}
	}

	d.readChan <- &output
}

func (d Device) Close() {
	d.conn.Close()
}