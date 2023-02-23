package msg

//TODO MOVE TO GANDALF
// Config : gandalf configs
type Config struct {
	MessageBase
	Target  string
	Command string // type of config
	Context map[string]interface{}
}

// NewConfig : Config constructor
// todo : passer une map pour gerer les valeurs optionnelles ?
func NewConfig(target string, command string, payload string) *Config {
	c := new(Config)
	c.InitMessageBase()

	c.Target = target
	c.Context = make(map[string]interface{})
	c.Command = command
	c.Payload = payload
	return c
}

// GetMessageType accessor
func (c Config) GetMessageType() string { return "config" }

// GetTarget :
func (c Config) GetTarget() string { return c.Target }

// GetCommand :
func (c Config) GetCommand() string { return c.Command }

// GetContext :
func (c Config) GetContext() map[string]interface{} { return c.Context }
