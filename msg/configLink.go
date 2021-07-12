package msg

// ConfigLink : Gandalf Socket config
type ConfigLink struct {
	MessageBase
	CommandName  string
	LogicalName  string
	ShosetType   string
	Address      string
	MyBrothers   []string
	YourBrothers []string
}

// func (c *ConfigLink) String() string {
// 	if c == nil {
// 		fmt.Printf("\nError : *ConfigLink.String : nil\n")
// 	}
// 	return fmt.Sprintf("[ CommandName: %s, LogicalName: %s, BindAddress: %s, Address: %s\n", c.CommandName, c.LogicalName, c.Address, c.Address)
// }

func NewCfgLink(address, lname, shosetType string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "link"
	c.Address = address
	c.LogicalName = lname
	c.ShosetType = shosetType
	return c
}

func NewCfgBrothers(myBrothers, yourBrothers []string) *ConfigLink {
	c := new(ConfigLink)
	c.InitMessageBase()
	c.CommandName = "brothers"
	c.MyBrothers = myBrothers
	c.YourBrothers = yourBrothers
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

// GetCommandName :
func (c ConfigLink) GetCommandName() string { return c.CommandName }

// GetConns :
func (c ConfigLink) GetMyBrothers() []string { return c.MyBrothers }

// GetBros :
func (c ConfigLink) GetYourBrothers() []string { return c.YourBrothers }
