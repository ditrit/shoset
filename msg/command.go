package msg

// Command : gandalf commands
type Command struct {
	MessageBase
	Target  string
	Command string // type of command
	Context map[string]interface{}
}

// NewCommand : Command constructor
// todo : passer une map pour gerer les valeurs optionnelles ?
func NewCommand(target string, command string, payload string) *Command {
	c := &Command{
		MessageBase: MessageBase{Payload: payload},
		Target:      target,
		Context:     make(map[string]interface{}),
		Command:     command,
	}
	c.InitMessageBase()
	return c
}

// GetMessageType accessor
func (c Command) GetMessageType() string { return "cmd" }

// GetTarget :
func (c Command) GetTarget() string { return c.Target }

// GetCommand :
func (c Command) GetCommand() string { return c.Command }

// GetContext :
func (c Command) GetContext() map[string]interface{} { return c.Context }
