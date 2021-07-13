package msg

import "fmt"

// ConfigJoin : Gandalf Socket config
type ConfigJoin struct {
	MessageBase
	CommandName string
	Address string
	Lname       string
	ShosetType  string
}

// GetMsgType accessor
func (c ConfigJoin) GetMsgType() string { return "cfgjoin" }

// GetAddress :
func (c ConfigJoin) GetAddress() string { return c.Address }

// GetCommandName :
func (c ConfigJoin) GetCommandName() string { return c.CommandName }

// GetName :
func (c ConfigJoin) GetLogicalName() string { return c.Lname }

// GetShosetType :
func (c ConfigJoin) GetShosetType() string { return c.ShosetType }

 // join/ok/member
func NewCfgJoin(address, lname, shosetType, commandName string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = commandName
	c.Address = address
	c.Lname = lname
	c.ShosetType = shosetType
	return c
}

func (c *ConfigJoin) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, BindAddress: %s\n", c.GetCommandName(), c.GetAddress())
}
