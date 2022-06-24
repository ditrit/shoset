package msg

// MessageBase base struct for messages containing raw data
type MessageBaseByte struct {
	MessageBase
	PayloadByte []byte
}

// InitMessageBase constructor
func (m *MessageBaseByte) InitMessageBase() {
	m.PayloadByte = []byte{}
	m.MessageBase.InitMessageBase()
}

//Many duplicate methods...
// GetUUID accessor
func (m MessageBaseByte) GetUUID() string {
	return m.UUID
}

func (m *MessageBaseByte) SetUUID(newUUID string) {
	m.UUID = newUUID
}

// GetTenant accessor
func (m MessageBaseByte) GetTenant() string {
	return m.Tenant
}

// GetToken accessor
func (m MessageBaseByte) GetToken() string {
	return m.Token
}

// GetTimestamp accessor
func (m MessageBaseByte) GetTimestamp() int64 {
	return m.Timestamp
}

// GetTimeout accessor
func (m MessageBaseByte) GetTimeout() int64 {
	return m.Timeout
}

// GetPayload accessor
func (m MessageBaseByte) GetPayload() string {
	return m.Payload
}

// GetMajor accessor
func (m MessageBaseByte) GetMajor() int8 {
	return m.Major
}

// GetMinor accessor
func (m MessageBaseByte) GetMinor() int8 {
	return m.Minor
}

// GetPayloadByte accessor
func (m MessageBaseByte) GetPayloadByte() []byte {
	return m.PayloadByte
}

type FileChunkMessage struct {
	MessageBaseByte
	FileName      string
	FileLen       int
	ChunkNumber   int
	ReferenceUUID string
}

// NewEvent : Event constructor
func NewfileChunkMessage(filename string, fileLen int, chunkNumber int, data []byte) *FileChunkMessage {
	fileChunk := new(FileChunkMessage)
	fileChunk.InitMessageBase()

	fileChunk.FileName = filename
	fileChunk.FileLen = fileLen
	fileChunk.ChunkNumber = chunkNumber
	fileChunk.Payload = ""
	fileChunk.PayloadByte = data
	// if val, ok := params["referenceUUID"]; ok {
	// 	fileChunk.ReferenceUUID = val
	// }
	return fileChunk
}

//Pas necessaire ?

// GetMsgType accessor
func (fileChunk FileChunkMessage) GetMsgType() string { return "fileChunk" }

// GetReferenceUUID :
func (fileChunk FileChunkMessage) GetReferenceUUID() string { return fileChunk.ReferenceUUID }

// GetFileName accessor
func (fileChunk FileChunkMessage) GetFileName() string { return fileChunk.FileName }

// GetFileLen accessor
func (fileChunk FileChunkMessage) GetFileLen() int { return fileChunk.FileLen }

// GetChunkNumber accessor
func (fileChunk FileChunkMessage) GetChunkNumber() int { return fileChunk.ChunkNumber }
