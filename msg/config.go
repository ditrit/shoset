package msg

// Config : Gandalf Socket config
type Config struct {
	MessageBase

	Info string
}

// NewConfig : Config constructor
func NewConfig(info string) *Config {
	c := new(Config)
	c.InitMessageBase()

	c.Info = info
	return c
}

// GetMsgType accessor
func (c Config) GetMsgType() string { return "cfg" }

// GetInfo :
func (c Config) GetInfo() string { return c.Info }
