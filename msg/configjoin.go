package msg

import "fmt"

// ConfigJoin : Gandalf Socket config
type ConfigJoin struct {
	MessageBase
	CommandName string
	BindAddress string
}

func (c *ConfigJoin) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, BindAddress: %s\n", c.CommandName, c.BindAddress)
}

// NewCfgJoin :
func NewCfgJoin(bindAddress string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "join"
	c.BindAddress = bindAddress
	// ajout nom logique et type de message afin de rajouter dans le protocole de join une condition vérifiant le même type et même nom logique pour établier un join
	return c
}

// NewCfgMember :
func NewCfgMember(bindAddress string) *ConfigJoin {
	c := new(ConfigJoin)
	c.InitMessageBase()
	c.CommandName = "member"
	c.BindAddress = bindAddress
	return c
}

// GetMsgType accessor
func (c ConfigJoin) GetMsgType() string { return "cfgjoin" }

// GetBindAddress :
func (c ConfigJoin) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c ConfigJoin) GetCommandName() string { return c.CommandName }
