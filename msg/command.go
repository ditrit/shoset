package msg

// Command : gandalf commands
type Command struct {
	MessageBase
	Target  string
	Command string
	Context map[string]interface{}
}

// NewCommand : Command constructor
// todo : passer une map pour gerer les valeurs optionnelles ?
func NewCommand(target string, command string, payload string) *Command {
	c := new(Command)
	c.InitMessageBase()

	c.Target = target
	c.Context = make(map[string]interface{})
	c.Command = command
	c.Payload = payload
	return c
}

// GetMsgType accessor
func (c Command) GetMsgType() string { return "cmd" }

// GetTarget :
func (c Command) GetTarget() string { return c.Target }

// GetCommand :
func (c Command) GetCommand() string { return c.Command }

// GetContext :
func (c Command) GetContext() map[string]interface{} { return c.Context }
