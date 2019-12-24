package msg

import "container/list"

// MBuffer :
type MBuffer struct {
	events   Queue
	commands Queue
	replies  Queue
	configs  Queue
}

// NewMBuffer :
func NewMBuffer() *MBuffer {
	b := new(MBuffer)
	b.Init()
	return b
}

// Init :
func (b *MBuffer) Init() {
	b.events.Init()
	b.commands.Init()
	b.replies.Init()
	b.configs.Init()
}

// PushEvent :
func (b *MBuffer) PushEvent(e Event) *list.Element {
	return b.events.Push(e)
}

// PushCommand :
func (b *MBuffer) PushCommand(e Command) *list.Element {
	return b.commands.Push(e)
}

// PushReply :
func (b *MBuffer) PushReply(e Reply) *list.Element {
	return b.replies.Push(e)
}

// PushConfig :
func (b *MBuffer) PushConfig(e Config) *list.Element {
	return b.configs.Push(e)
}

// GetEvent :
func (b *MBuffer) GetEvent(ctx *list.Element) Event {
	return b.events.Get(ctx).(Event)
}

// GetCommand :
func (b *MBuffer) GetCommand(ctx *list.Element) Command {
	return b.commands.Get(ctx).(Command)
}

// GetReply :
func (b *MBuffer) GetReply(ctx *list.Element) Reply {
	return b.replies.Get(ctx).(Reply)
}

// GetConfig :
func (b *MBuffer) GetConfig(ctx *list.Element) Config {
	return b.configs.Get(ctx).(Config)
}

// RemoveEvent :
func (b *MBuffer) RemoveEvent(evt Event) {
	b.events.Remove(evt)
}

// RemoveCommand :
func (b *MBuffer) RemoveCommand(evt Command) {
	b.commands.Remove(evt)
}

// RemoveReply :
func (b *MBuffer) RemoveReply(evt Reply) {
	b.replies.Remove(evt)
}

// RemoveConfig :
func (b *MBuffer) RemoveConfig(evt Config) {
	b.configs.Remove(evt)
}
