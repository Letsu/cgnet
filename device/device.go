package device

import "context"

type Cli struct {
	data string
}

type Facts struct {
	DeviceType string
	Hostname   string
	Version    string
	Uptime     string
}

type Device interface {
	// Connects to the device. Login and privilage escalation should be done after this
	// The device should be ready to accept commands.
	Connect(ctx context.Context) error
	// Cli executes a command on the device and returns the output
	// Todo how to handle non cli methodes
	Cli([]string) (map[string]Cli, error)
	// Facts returens the basic informations from a device needed to identify it
	Facts() (Facts, error)
	// Close
	Close() error
}
