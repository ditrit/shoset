package msg

// ConfigProtocol : Gandalf Socket config
type ConfigProtocol struct {
	MessageBase
	CommandName  string
	LogicalName  string
	ShosetType   string
	Address      string
	MyBrothers   []string
	YourBrothers []string
}

// for link and join
func NewCfg(address, lName, shosetType, commandName string) *ConfigProtocol {
	c := new(ConfigProtocol)
	c.InitMessageBase()
	c.CommandName = commandName
	c.Address = address
	c.LogicalName = lName
	c.ShosetType = shosetType
	return c
}

// for link
func NewCfgBrothers(myBrothers, yourBrothers []string, lName, commandName, shosetType string) *ConfigProtocol {
	c := new(ConfigProtocol)
	c.InitMessageBase()
	c.CommandName = commandName
	c.MyBrothers = myBrothers
	c.YourBrothers = yourBrothers
	c.LogicalName = lName
	c.ShosetType = shosetType
	return c
}

// GetMsgType accessor
func (c ConfigProtocol) GetMsgType() string {
	switch c.GetCommandName() {
	case "join":
		return "cfgjoin"
	case "aknowledge_join":
		return "cfgjoin"
	case "unaknowledge_join":
		return "cfgjoin"
	case "member":
		return "cfgjoin"
	case "link":
		return "cfglink"
	case "brothers":
		return "cfglink"
	case "bye":
		return "case note treated yet"
	}
	return "Wrong input protocolType"
}

// GetLogicalName :
func (c ConfigProtocol) GetLogicalName() string { return c.LogicalName }

// GetShosetType :
func (c ConfigProtocol) GetShosetType() string { return c.ShosetType }

// GetAddress :
func (c ConfigProtocol) GetAddress() string { return c.Address }

// GetCommandName :
func (c ConfigProtocol) GetCommandName() string { return c.CommandName }

// GetConns :
func (c ConfigProtocol) GetMyBrothers() []string { return c.MyBrothers }

// GetBros :
func (c ConfigProtocol) GetYourBrothers() []string { return c.YourBrothers }
