package msg

import "fmt"

// ConfigJoin : Gandalf Socket config
type ConfigJoin struct {
	MessageBase
	CommandName string
	BindAddress string
	Lname       string
	ShosetType  string
}

// GetMsgType accessor
func (c ConfigJoin) GetMsgType() string { return "cfgjoin" }

// GetBindAddress :
func (c ConfigJoin) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c ConfigJoin) GetCommandName() string { return c.CommandName }

// GetName :
func (c ConfigJoin) GetName() string { return c.Lname }

// GetShosetType :
func (c ConfigJoin) GetShosetType() string { return c.ShosetType }

// NewCfgJoin :
func NewCfgJoin(bindAddress, Lname, ShosetType string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "join"
	c.BindAddress = bindAddress
	c.Lname = Lname
	c.ShosetType = ShosetType
	return c
}

func NewCfgJoinOk() *ConfigJoin {
	c := new(ConfigJoin)
	c.CommandName = "ok"
	return c
}

// NewCfgMember :
func NewCfgMember(bindAddress, Lname, ShosetType string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "member"
	c.BindAddress = bindAddress
	c.Lname = Lname
	c.ShosetType = ShosetType
	return c
}

func (c *ConfigJoin) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, BindAddress: %s\n", c.CommandName, c.BindAddress)
}
