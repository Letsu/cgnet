package main

import (
	"github.com/letsu/goCisco"
	"log"
)

func main() {
	dev := goCisco.Device{
		Ip:         "192.168.0.53",
		Username:   "",
		Password:   "",
		Enable:     "",
		DeviceType: "",
		ConnType:   "telnet",
	}

	err := dev.Open()
	log.Println(err)
}
