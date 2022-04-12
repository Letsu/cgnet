package cgnet

import (
	"bufio"
	"net"
	"time"
)

func (d *Device) connectTelnet() error {
	var err error
	if d.Port == "" {
		d.Port = "23"
	}
	d.telnetClient, err = net.DialTimeout("tcp", d.Ip+":"+d.Port, 10*time.Second)
	if err != nil {
		return err
	}
	d.stdout = bufio.NewReader(d.telnetClient)
	d.stdin = d.telnetClient
	d.readChan = make(chan *string, 20)

	err = d.login()
	if err != nil {
		return err
	}

	d.Exec("terminal length 0")

	return nil
}
