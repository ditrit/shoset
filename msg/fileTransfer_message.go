package msg

type FileChunkMessage struct {
	MessageBase
	FileName      string
	FileLen       int
	ChunkNumber   int
	ReferenceUUID string
}

// NewFileChunkMessage : FileChunkMessage constructor
func NewFileChunkMessage(filename string, fileLen int, chunkNumber int, data []byte) *FileChunkMessage {
	fileChunk := new(FileChunkMessage)
	fileChunk.InitMessageBase()

	fileChunk.FileName = filename
	fileChunk.FileLen = fileLen
	fileChunk.ChunkNumber = chunkNumber
	fileChunk.PayloadByte = data

	// ??
	// if val, ok := params["referenceUUID"]; ok {
	// 	fileChunk.ReferenceUUID = val
	// }
	return fileChunk
}

// GetMsgType accessor
func (fileChunk FileChunkMessage) GetMsgType() string { return "fileChunk" }

// GetReferenceUUID :
func (fileChunk FileChunkMessage) GetReferenceUUID() string { return fileChunk.ReferenceUUID }

//Specific to FileChunkMessage
// GetFileName accessor
func (fileChunk FileChunkMessage) GetFileName() string { return fileChunk.FileName }

// GetFileLen accessor
func (fileChunk FileChunkMessage) GetFileLen() int { return fileChunk.FileLen }

// GetChunkNumber accessor
func (fileChunk FileChunkMessage) GetChunkNumber() int { return fileChunk.ChunkNumber }
