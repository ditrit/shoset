package msg

import "fmt"

// Config : Gandalf Socket config
type Config struct {
	MessageBase
	CommandName string
	LogicalName string
	BindAddress string
	Address     string
	Conns       []string
}

func (c *Config) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, LogicalName: %s, BindAddress: %s, Address: %s, Conns : %#v\n", c.CommandName, c.LogicalName, c.BindAddress, c.Address, c.Conns)
}

// NewHandshake : Config constructor
func NewHandshake(bindAddress, logicalName string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "handshake"
	c.LogicalName = logicalName
	c.BindAddress = bindAddress
	return c
}

// NewNameBrothers : Config constructor
func NewNameBrothers(brothers []string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "nameBrothers"
	c.Conns = brothers
	return c
}

// NewConns : Config constructor
func NewConns(dir string, conns []string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = dir
	c.Conns = conns
	return c
}

// NewInstance : Config constructor
func NewInstance(address string, logicalName string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "newInstance"
	c.LogicalName = logicalName
	c.Address = address
	return c
}

// NewConnectTo : Config constructor
func NewConnectTo(address string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "connectTo"
	c.Address = address
	return c
}

// GetMsgType accessor
func (c Config) GetMsgType() string { return "cfg" }

// GetLogicalName :
func (c Config) GetLogicalName() string { return c.LogicalName }

// GetAddress :
func (c Config) GetAddress() string { return c.Address }

// GetBindAddress :
func (c Config) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c Config) GetCommandName() string { return c.CommandName }

// GetConns :
func (c Config) GetConns() []string { return c.Conns }
