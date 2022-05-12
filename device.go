package cgnet

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type Device struct {
	Ip           string
	Port         string
	Username     string
	Password     string
	Enable       string
	DeviceType   string
	ConnType     string
	telnetClient net.Conn
	sshClient    *ssh.Client
	sshSession   *ssh.Session
	stdin        io.Writer
	stdout       io.Reader
	readChan     chan *string
	prompt       string
}

var (
	ErrUnknownCommand = errors.New("unknown or invalid command")
	ErrAuthFailed     = errors.New("authentication failed")
	ErrNoPrompt       = errors.New("no return of prompt after command")
	ErrUnsupported    = errors.New("unsupported connection type")

	// Timeout for waiting for a prompt
	Timeout = time.Second * 5
	// Log all to stdout of conso
	ShowLog = false

	// Connection types
	Telnet = "telnet"
	SSH    = "ssh"
)

// Open the connection to the device
func (d *Device) Open() error {
	switch d.ConnType {
	case Telnet:
		err := d.connectTelnet()
		if err != nil {
			return err
		}
		break
	case SSH:
		err := d.connectSSH()
		if err != nil {
			return err
		}
		break
	default:
		return ErrUnsupported
	}

	return nil
}

func (d *Device) getPrompt() *regexp.Regexp {
	if len(d.prompt) > 10 {
		d.prompt = d.prompt[:10]
	}
	return regexp.MustCompile(d.prompt + "[[:alnum:]-_]*[\\#>]")
}

// Exec2 executes a command on the device and without returning the output
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
	match, _ = regexp.MatchString(d.getPrompt().String(), text)
	//Login
	for !match {
		switch {
		case strings.Contains(text, "timeout"):
			return errors.New("timeout")
		case strings.Contains(text, "sername:"):
			_, err := io.WriteString(d.stdin, d.Username+"\n")
			if err != nil {
				return err
			}
			break
		case strings.Contains(text, "assword:"):
			_, err := io.WriteString(d.stdin, d.Password+"\n")
			if err != nil {
				return err
			}
			break

		case strings.Contains(text, "Authentication failed"):
			return ErrAuthFailed
		default:
			break
		}
		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
		match, _ = regexp.MatchString(d.getPrompt().String(), text)
	}
	d.prompt = d.getPrompt().FindString(text)

	// Enable
	enabled := !strings.Contains(d.prompt, ">")
	if d.Enable == "" {
		enabled = true
	}

	if !enabled {
		_, err := io.WriteString(d.stdin, "enable\n")
		if err != nil {
			return err
		}
		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
	}
	for !enabled {
		switch {
		case strings.Contains(text, "assword:"):
			_, err := io.WriteString(d.stdin, d.Enable+"\n")
			if err != nil {
				return err
			}
			break
		default:
			break
		}

		n, _ = d.stdout.Read(buf)
		text = string(buf[:n])
		enabled, _ = regexp.MatchString("[[:alnum:]-_]*[\\#]", text)
	}

	d.prompt = strings.Replace(d.prompt, ">", "", -1)
	d.prompt = strings.Replace(d.prompt, "#", "", -1)

	return nil
}

// Exec executes a command on the device and returns the output
func (d *Device) Exec(cmd ...string) (string, error) {
	go d.reader(cmd...)
	_, err := io.WriteString(d.stdin, fmt.Sprint(strings.Join(cmd, ""), "\n"))
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
		case <-time.After(Timeout):
			return "", ErrNoPrompt
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

		if ShowLog {
			fmt.Println(string(buf[:n]))
		}
	}

	d.readChan <- &output
}

// Close the connection to the device
func (d Device) Close() error {
	close(d.readChan)
	if d.ConnType == "telnet" {
		return d.telnetClient.Close()
	} else if d.ConnType == "ssh" {
		d.sshSession.Close()
		return d.sshClient.Close()
	}

	return nil
}
