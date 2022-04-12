[![Go Reference](https://pkg.go.dev/badge/github.com/Letsu/cgnet.svg)](https://pkg.go.dev/github.com/Letsu/cgnet)[![Go Report Card](https://goreportcard.com/badge/github.com/Letsu/cgnet)](https://goreportcard.com/report/github.com/Letsu/cgnet)
# CG - NET

CG-Net is a simple netmiko like package for go to manage cisco devices.

Currently, It's possible to connect to devices using telnet and ssh. 

Tested on various cisco catalyst switches, ASRs and ISRs. Other devices should also be possible (:

Installation
------------
``` sh
go get github.com/Letsu/cgnet
```

Example
-------
Get Version and configure a loopback interface
```go
package main

import (
	"fmt"
	"github.com/Letsu/cgnet"
)

func main() {
	d := cgnet.Device{
		Ip:       "10.10.10.10",
		Username: "cisco",
		Password: "cisco",
		Enable:   "cisco",
		ConnType: cgnet.SSH,
	}

	err := d.Open()
	defer d.Close()
	if err != nil {
		panic(err)
	}

	ver, err := d.Exec("sh version")
	if err != nil {
		panic(err)
	}
	fmt.Println(ver)

	cmds := []string{"interface loopback10", "ip address 10.10.10.11 255.255.255.255", "no shut"}
	err = d.Configure(cmds)
	if err != nil {
		panic(err)
	}
}
```
