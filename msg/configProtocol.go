package msg

// ConfigProtocol : config handler for each protocol of Shoset.
type ConfigProtocol struct {
	MessageBase
	CommandName  string   // type of config
	LogicalName  string   // logical name of the Shoset
	ShosetType   string   // type of the Shoset
	BindAddress  string   // bindAddress of the Shoset
	MyBrothers   []string // list of ipAddresses of same Shoset logical name
	YourBrothers []string // list of ipAddresses of other Shoset logical name
}

// NewConfigProtocol is a simple config for each protocol.
func NewConfigProtocol(address, lName, shosetType, commandName string) *ConfigProtocol {
	c := new(ConfigProtocol)
	c.InitMessageBase()
	c.CommandName = commandName
	c.BindAddress = address
	c.LogicalName = lName
	c.ShosetType = shosetType
	return c
}

// NewConfigBrothersProtocol is dedicated to Link protocol.*
//
func NewConfigBrothersProtocol(myBrothers, yourBrothers []string, commandName, lName, shosetType string) *ConfigProtocol {
	c := new(ConfigProtocol)
	c.InitMessageBase()
	c.CommandName = commandName
	c.MyBrothers = myBrothers
	c.YourBrothers = yourBrothers
	c.LogicalName = lName
	c.ShosetType = shosetType
	return c
}

// GetMessageType returns MessageType from ConfigProtocol.
func (c ConfigProtocol) GetMessageType() string {
	switch c.GetCommandName() {
	case "join":
		return "cfgjoin"
	case "acknowledge_join":
		return "cfgjoin"
	case "member":
		return "cfgjoin"
	case "link":
		return "cfglink"
	case "brothers":
		return "cfglink"
	case "acknowledge_link":
		return "cfglink"
	case "bye":
		return "cfgbye"
	case "delete":
		return "cfgbye"
	}
	return "Wrong input protocolType"
}

// GetLogicalName returns LogicalName from ConfigProtocol.
func (c ConfigProtocol) GetLogicalName() string { return c.LogicalName }

// GetShosetType returns ShosetType from ConfigProtocol.
func (c ConfigProtocol) GetShosetType() string { return c.ShosetType }

// GetAddress returns BindAddress from ConfigProtocol.
func (c ConfigProtocol) GetAddress() string { return c.BindAddress }

// GetCommandName returns CommandName from ConfigProtocol.
func (c ConfigProtocol) GetCommandName() string { return c.CommandName }

// GetMyBrothers returns MyBrothers from ConfigProtocol.
func (c ConfigProtocol) GetMyBrothers() []string { return c.MyBrothers }

// GetYourBrothers returns YourBrothers from ConfigProtocol.
func (c ConfigProtocol) GetYourBrothers() []string { return c.YourBrothers }
