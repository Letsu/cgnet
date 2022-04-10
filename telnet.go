package goCisco

import (
	"bufio"
	"net"
)

func (d *Device) connectTelnet() error {
	var err error
	if d.Port == "" {
		d.Port = "23"
	}
	d.conn, err = net.Dial("tcp", d.Ip+":"+d.Port)
	if err != nil {
		return err
	}
	d.stdout = bufio.NewReader(d.conn)
	d.stdin = bufio.NewWriter(d.conn)
	d.readChan = make(chan *string, 20)

	err = d.login()
	if err != nil {
		return err
	}

	d.Exec("terminal length 0")

	return nil
}
