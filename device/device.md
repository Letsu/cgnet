# Device
The Device interface implemtation is the low level implementiotion of a connection to a network device. 
It should handle
- Connection to the device and handeling the session. 
- Login and/or authentification
- Privilage escalation
- Executing commands and returning the output
- Parsing the reurn to a plain text format or already parsed into some data structure if already returned
- Reading the basic facts from the device   
- All vendor specific extra steps requiert to communicate

## Creating a device driver
Each device class should be its own package, while different methode to connect to a device(like ssh, telnet, net/restconf, snmp) can be implemented as different types in the package. 
All devices should implement the Device interface.

### Conneciong

### Cli

### Close