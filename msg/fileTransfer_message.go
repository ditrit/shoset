package msg

type FileChunkMessage struct {
	MessageBaseByte
	FileName      string
	FileLen       int
	ChunkNumber int
	ReferenceUUID string
}

// NewEvent : Event constructor
func NewfileChunkMessage(filename string, fileLen int, chunkNumber int, data []byte) *FileChunkMessage {
	fileChunk := new(FileChunkMessage)
	fileChunk.InitMessageBase()

	fileChunk.FileName = filename
	fileChunk.FileLen  = fileLen
	fileChunk.ChunkNumber = chunkNumber
	fileChunk.Payload = ""
	fileChunk.PayloadByte= data
	// if val, ok := params["referenceUUID"]; ok {
	// 	fileChunk.ReferenceUUID = val
	// }
	return fileChunk
}

// GetMsgType accessor
func (fileChunk FileChunkMessage) GetMsgType() string { return "fileChunk" }

// GetReferenceUUID :
func (fileChunk FileChunkMessage) GetReferenceUUID() string { return fileChunk.ReferenceUUID }
