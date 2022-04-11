package cgnet

import (
	"golang.org/x/crypto/ssh"
)

func (d *Device) connectSSH() error {
	var err error
	if d.Port == "" {
		d.Port = "22"
	}
	sshConf := ssh.ClientConfig{
		User: d.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(d.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	d.sshClient, err = ssh.Dial("tcp", d.Ip+":"+d.Port, &sshConf)
	if err != nil {
		return err
	}
	d.sshSession, err = d.sshClient.NewSession()
	if err != nil {
		return err
	}
	d.stdin, err = d.sshSession.StdinPipe()
	if err != nil {
		return err
	}
	d.stdout, err = d.sshSession.StdoutPipe()
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0, // disable echoing
		ssh.OCRNL:         0,
		ssh.TTY_OP_ISPEED: 38400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 38400, // output speed = 14.4kbaud
	}
	d.sshSession.RequestPty("vt100", 0, 2000, modes)
	d.sshSession.Shell()

	d.readChan = make(chan *string, 20)

	err = d.login()
	if err != nil {
		return err
	}

	_, err = d.Exec("terminal length 0")
	if err != nil {
		return err
	}

	return nil
}
