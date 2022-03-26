package goCisco

import (
	"fmt"
)

func (d Device) Configure(cmds []string) error {

	err := d.Exec2("conf t")
	if err != nil {
		return err
	}

	err = nil
	for _, cmd := range cmds {
		if d.Exec2(cmd) != nil {
			err = fmt.Errorf("error on command %s. aborting, %s", cmd, err.Error())
			break
		}
	}
	if err != nil {
		return err
	}

	return nil
}
