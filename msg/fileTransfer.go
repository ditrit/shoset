package msg

type FileStateMessage struct {
	UUID          string
	Name          string
	Hash          string
	Version       int
	Path          string
	LastOperation OperationMessage
}

type OperationMessage struct {
	Name    string
	File    string
	NewFile string
	Version int
	Hash    string
}

type FileMessage struct {
	MessageBase
	MessageName string

	// file description
	FileName string
	FileHash string
	FileUUID string

	// for file info message
	FileSize      int64
	FileVersion   int
	FilePath      string
	FileOperation OperationMessage
	FileHashMap   map[int]string

	// for library message
	Library []FileStateMessage
}

func (fileMessage FileMessage) GetMessageType() string { return "file" }
