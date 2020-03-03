package msg

import "fmt"

// ConfigLink : Gandalf Socket config
type ConfigLink struct {
	MessageBase
	CommandName string
	LogicalName string
	ShosetType  string
	BindAddress string
	Address     string
	Conns       []string
	Bros        []string
}

func (c *ConfigLink) String() string {
	if c == nil {
		fmt.Printf("\nError : *ConfigLink.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, LogicalName: %s, BindAddress: %s, Address: %s, Conns : %#v\n", c.CommandName, c.LogicalName, c.BindAddress, c.Address, c.Conns)
}

// NewHandshake : ConfigLink constructor
func NewHandshake(bindAddress, logicalName, ShosetType string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "handshake"
	c.LogicalName = logicalName
	c.BindAddress = bindAddress
	c.ShosetType = ShosetType
	return c
}

// NewNameBrothers : ConfigLink constructor
func NewNameBrothers(brothers []string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "nameBrothers"
	c.Conns = brothers
	return c
}

// NewConns : ConfigLink constructor
func NewConns(dir string, conns []string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = dir
	c.Conns = conns
	return c
}

// NewInstance : ConfigLink constructor
func NewInstance(address, logicalName, ShosetType string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "newInstance"
	c.LogicalName = logicalName
	c.ShosetType = ShosetType
	c.Address = address
	return c
}

// NewConnectTo : ConfigLink constructor
func NewConnectTo(address string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "connectTo"
	c.Address = address
	return c
}

// GetMsgType accessor
func (c ConfigLink) GetMsgType() string { return "cfglink" }

// GetLogicalName :
func (c ConfigLink) GetLogicalName() string { return c.LogicalName }

// GetShosetType :
func (c ConfigLink) GetShosetType() string { return c.ShosetType }

// GetAddress :
func (c ConfigLink) GetAddress() string { return c.Address }

// GetBindAddress :
func (c ConfigLink) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c ConfigLink) GetCommandName() string { return c.CommandName }

// GetConns :
func (c ConfigLink) GetConns() []string { return c.Conns }

// GetBros :
func (c ConfigLink) GetBros() []string { return c.Bros }
