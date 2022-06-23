package files

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

//Experimental prototype using events

// Get ?? :

// HandleFileTranfer :
func (fileTransfer *FileTransfer) HandleTransfer() {
	// //File infos :
	// fileInfo := fileTransfer.file.Name + "," + fmt.Sprint(len(fileTransfer.file.Data))
	// event := msg.NewEventClassic("fileTransfer", "fileTransferStart", fileInfo)

	// //Send file infos at every destinations
	// for conn, _ := range fileTransfer.requestedChunks {
	// 	conn.SendMessage(event)
	// }

	fileChunk := []byte{}

	//Send requested chunks :
	for conn, chRqByConn := range fileTransfer.requestedChunks {
		//fmt.Println("conn (HandleTransfer) : ", conn)
		//fmt.Println("chRqByConn(HandleTransfer) : ", chRqByConn)
		for _, chunk := range chRqByConn {
			fmt.Println("chunk (HandleTransfer) : ", chunk)
			//Get data :
			lenFile := len(fileTransfer.file.Data)
			if (chunk+1)*chunkSize < lenFile {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : (chunk+1)*chunkSize]
			} else {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : lenFile]
			}

			//Generate message :
			message := msg.NewfileChunkMessage(fileTransfer.file.Name, lenFile, chunk, fileChunk)

			fmt.Println("message (HandleTransfer) : ", message)

			err := conn.SendMessage(message)

			time.Sleep(10 * time.Millisecond)

			if err != nil {
				fmt.Println("sendChunk : ", err)
			}
		}
	}

	//Check if no other chunks are requested :
}

func requestFile(name string) {
	//Request File infos :

	//Create File :

	//Request chunks ?? :

	//Start reception :
}

// SendChunk :
func (file *File) sendChunk(destination *shoset.ShosetConn, chunkNumber int) {
	//Create event :
	var fileChunk string
	if (chunkNumber+1)*chunkSize < len(file.Data) {
		fileChunk = fmt.Sprint(chunkNumber) + "," + fmt.Sprint(file.Data[chunkNumber*chunkSize:(chunkNumber+1)*chunkSize])
	} else {
		fileChunk = fmt.Sprint(chunkNumber) + "," + fmt.Sprint(file.Data[chunkNumber*chunkSize:len(file.Data)])
	}

	event := msg.NewEventClassic("fileTransfer", "fileTransferStart", fileChunk)

	fmt.Println("event (sendChunk)", event)

	//Send event :
	time.Sleep(10 * time.Millisecond)
	err := destination.SendMessage(event)

	if err != nil {
		fmt.Println("sendChunk : ", err)
	}
}

// WaitFile :
func (transfer *FileTransfer) WaitFile(c *shoset.Shoset) *File {
	file := NewEmptyFile()

	file.m.Lock()

	iterator := msg.NewIterator(c.Queue["fileChunk"])

	//Get File info in first message :
	chunk_rc := c.Wait("fileChunk", map[string]string{}, 5, iterator)
	for chunk_rc != nil {
		fmt.Println("chunk_rc (WaitFile) : ",chunk_rc)
		chunk_rc = c.Wait("fileChunk", map[string]string{}, 5, iterator)	
	}

	//payload := chunk_rc.GetPayload()
	//data := chunk_rc.GetPayloadByte()

	// dataString := strings.Split(payload, ",")

	// fmt.Println("dataString (WaitFile)", dataString)

	// file.Name =dataString[0]
	// fileLen,err :=strconv.Atoi(dataString[1])

	// fmt.Println("file.Name (WaitFile)", file.Name)
	// fmt.Println("fileLen (WaitFile)", fileLen)

	// if err != nil {
	// 	fmt.Println("WaitFile : ", err)
	// }

	// for i := 0; i < ((fileLen/chunkSize)+1); i++ {
	// 	event_rc = c.Wait("evt", map[string]string{"topic": "fileTransfer", "event": "fileTransferStart"}, 5, iterator)
	// 	payload = event_rc.GetPayload()

	// 	fmt.Println("payload (WaitFile)", payload)

	// 	dataString = strings.Split(payload, ",")

	// 	//Convert payload back to string and Add data to file
	// 	fmt.Println("strings.Split()",strings.Split(strings.Trim(dataString[1], "[]"), " "))

	// 	for _, ps := range strings.Split(strings.Trim(dataString[1], "[]"), " ") {
	// 		pi,_ := strconv.Atoi(ps)
	// 		file.Data = append(file.Data,byte(pi))
	// 	}
	// }
	file.m.Unlock()
	file.Status = "ready"
	return file
}

// WaitChunk :
// func (file *File) WaitChunk(c shoset.Shoset, iterator *msg.Iterator) []byte {
// 	//Find chunk number :

// 	event_rc := c.Wait("evt", map[string]string{"topic": "fileTransfer", "event": "fileTransferStart"}, 5, iterator) //Ne consomme pas les messages
// 	payload := event_rc.GetPayload()
// 	shoset.Log("\nevent_rc (Payload) (WaitChunk) : " + payload)

// 	//Add to file data :
// 	data := []byte{}

// 	return data
// }
