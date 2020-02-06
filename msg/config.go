package msg

import "fmt"

// Config : Gandalf Socket config
type Config struct {
	MessageBase
	CommandName string
	LogicalName string
	ShosetType  string
	BindAddress string
	Address     string
	Conns       []string
	Bros        []string
}

func (c *Config) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, LogicalName: %s, BindAddress: %s, Address: %s, Conns : %#v\n", c.CommandName, c.LogicalName, c.BindAddress, c.Address, c.Conns)
}

// NewCfgJoin :
func NewCfgJoin(bindAddress string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "join"
	c.BindAddress = bindAddress
	return c
}

// NewCfgMember :
func NewCfgMember(bindAddress string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "member"
	c.BindAddress = bindAddress
	return c
}

// NewHandshake : Config constructor
func NewHandshake(bindAddress, logicalName, ShosetType string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "handshake"
	c.LogicalName = logicalName
	c.BindAddress = bindAddress
	c.ShosetType = ShosetType
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
func NewInstance(address, logicalName, ShosetType string) *Config {
	c := new(Config)
	c.InitMessageBase()
	c.CommandName = "newInstance"
	c.LogicalName = logicalName
	c.ShosetType = ShosetType
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

// GetShosetType :
func (c Config) GetShosetType() string { return c.ShosetType }

// GetAddress :
func (c Config) GetAddress() string { return c.Address }

// GetBindAddress :
func (c Config) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c Config) GetCommandName() string { return c.CommandName }

// GetConns :
func (c Config) GetConns() []string { return c.Conns }

// GetBros :
func (c Config) GetBros() []string { return c.Bros }
