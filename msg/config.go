package msg

// Config : Gandalf Socket config
type Config struct {
	MessageBase

	LogicalName string
}

// NewConfig : Config constructor
func NewConfig(logicalName string) *Config {
	c := new(Config)
	c.InitMessageBase()

	c.LogicalName = logicalName
	return c
}

// GetMsgType accessor
func (c Config) GetMsgType() string { return "cfg" }

// GetLogicalName :
func (c Config) GetLogicalName() string { return c.LogicalName }
