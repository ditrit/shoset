package msg

import "fmt"

// ConfigBye : Gandalf Socket config
type ConfigBye struct {
	MessageBase
	CommandName string
	BindAddress string
	LogicalName string
}

func (c *ConfigBye) String() string {
	if c == nil {
		fmt.Printf("\nError : *Config.String : nil\n")
	}
	return fmt.Sprintf("[ CommandName: %s, BindAddress: %s\n", c.CommandName, c.BindAddress)
}

// NewCfgBye :
func NewCfgBye(bindAddress, lName string) *ConfigBye {
	c := new(ConfigBye)
	c.InitMessageBase()
	c.BindAddress = bindAddress
	c.CommandName = "bye"
	c.LogicalName = lName
	return c
}

// NewCfgByeOk :
func NewCfgByeOk(bindAddress string) *ConfigBye {
	c := new(ConfigBye)
	c.InitMessageBase()
	c.BindAddress = bindAddress
	c.CommandName = "bye_ok"
	return c
}

// NewCfgByeBro :
func NewCfgByeBro(bindAddress string) *ConfigBye {
	c := new(ConfigBye)
	c.InitMessageBase()
	c.BindAddress = bindAddress
	c.CommandName = "bye_bro"
	return c
}

// GetMsgType accessor
func (c ConfigBye) GetMsgType() string { return "cfgbye" }

// GetBindAddress :
func (c ConfigBye) GetBindAddress() string { return c.BindAddress }

// GetCommandName :
func (c ConfigBye) GetCommandName() string { return c.CommandName }

// GetLogicalName :
func (c ConfigBye) GetLogicalName() string { return c.LogicalName }
