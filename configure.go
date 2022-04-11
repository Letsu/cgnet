package cgnet

import (
	"errors"
)

// Configure a silce of commands on the device. The commands executed automatically in the 'configure terminal' mode.
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
