package ciscoIOS

import (
	"context"

	"github.com/letsu/cgnet/device"
)

type IosSSH struct {
	Hostname string
	Username string
	Password string
	Enable   string
}

func (d *IosSSH) Connect(ctx context.Context) error {
	return nil
}

func (d *IosSSH) Cli([]string) (map[string]device.Cli, error) {
	return nil, nil
}

func (d *IosSSH) Facts() (device.Facts, error) {
	return device.Facts{}, nil
}

func (d *IosSSH) Close() error {
	return nil
}
