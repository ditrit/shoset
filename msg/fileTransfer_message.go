package msg

type FileChunkMessage struct {
	MessageBase
	FileName      string
	FileLen       int
	ChunkNumber   int
	ReferenceUUID string
}

// type handledFiles struct {
// 	HandledFilesList []string
// 	m                sync.Mutex
// }

// var HandledFiles1 handledFiles

// func CheckIfFileIsHandled(fileName string) bool {
// 	HandledFiles1.m.Lock()
// 	defer HandledFiles1.m.Unlock()
// 	for _, a := range HandledFiles1.HandledFilesList {
// 		if a == fileName {
// 			return true
// 		}
// 	}
// 	return false
// }

// func DeleteFromFileIsHandled(fileName string) {
// 	HandledFiles1.m.Lock()
// 	defer HandledFiles1.m.Unlock()
// 	for i, a := range HandledFiles1.HandledFilesList {
// 		if a == fileName {
// 			HandledFiles1.HandledFilesList = append(HandledFiles1.HandledFilesList[:i], HandledFiles1.HandledFilesList[i+1:]...)
// 		}
// 	}
// }

// NewFileChunkMessage : FileChunkMessage constructor
func NewFileChunkMessage(filename string, fileLen int, chunkNumber int, data []byte) *FileChunkMessage {
	fileChunk := new(FileChunkMessage)
	fileChunk.InitMessageBase()

	fileChunk.FileName = filename
	fileChunk.FileLen = fileLen
	fileChunk.ChunkNumber = chunkNumber
	//fileChunk.Payload = ""
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
