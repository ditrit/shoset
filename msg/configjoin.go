package msg

import "fmt"

// ConfigJoin : Gandalf Socket config
type ConfigJoin struct {
	MessageBase
	CommandName string
	BindAddress string
	name        string
	ShosetType  string
}

// GetMsgType accessor
func (c ConfigJoin) GetMsgType() string { return "cfgjoin" }

// GetBindAddress :
func (c ConfigJoin) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c ConfigJoin) GetCommandName() string { return c.CommandName }

// GetName :
func (c ConfigJoin) GetName() string { return c.name }

// GetShosetType :
func (c ConfigJoin) GetShosetType() string { return c.ShosetType }

// NewCfgJoin :
func NewCfgJoin(bindAddress string, name string, ShosetType string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "join"
	c.BindAddress = bindAddress
	c.name = name
	c.ShosetType = ShosetType
	return c
}

// NewCfgMember :
func NewCfgMember(bindAddress string, name string, ShosetType string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "member"
	c.BindAddress = bindAddress
	c.name = name
	c.ShosetType = ShosetType
	return c
}

func (c *ConfigJoin) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, BindAddress: %s\n", c.CommandName, c.BindAddress)
}
