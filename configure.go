package cgnet

import (
	"errors"
)

func (d Device) Configure(cmds []string) error {

	err := d.Exec2("conf t")
	if err != nil {
		return err
	}

	err = nil
	for _, cmd := range cmds {
		err = d.Exec2(cmd)
		if err != nil {
			err = errors.New("error on command " + cmd + ". aborting, " + err.Error())
			break
		}
	}
	d.Exec2("end")
	if err != nil {
		return err
	}

	return nil
}
