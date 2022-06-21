package files

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

//Experimental prototype using events

// Get ?? :

// HandleFileTranfer :
func (fileTransfer *FileTransfer) HandleTransfer() {
	//File infos :
	fileInfo := fileTransfer.file.Name + "," + fmt.Sprint(len(fileTransfer.file.Data))
	event := msg.NewEventClassic("fileTransfer", "fileTransferStart", fileInfo)

	//Send file infos at every destinations
	for conn, _ := range fileTransfer.requestedChunks {
		conn.SendMessage(event)
	}

	//Send requested chunks :
	for conn, chRqByConn := range fileTransfer.requestedChunks {
		//fmt.Println("conn (HandleTransfer) : ", conn)
		//fmt.Println("chRqByConn(HandleTransfer) : ", chRqByConn)
		for _, chunk := range chRqByConn {
			//fmt.Println("chunk (HandleTransfer) : ", chunk)
			fileTransfer.file.sendChunk(conn, chunk) //Send one of the requested chunks on the requesting conn
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

	iterator := msg.NewIterator(c.Queue["evt"])

	//Get File info in first message :
	event_rc := c.Wait("evt", map[string]string{"topic": "fileTransfer", "event": "fileTransferStart"}, 5, iterator)
	payload := event_rc.GetPayload()

	dataString := strings.Split(payload, ",")

	fmt.Println("dataString (WaitFile)", dataString)

	file.Name = dataString[0]
	fileLen,err := strconv.Atoi(dataString[1]) //int(dataString[1])

	fmt.Println("file.Name (WaitFile)", file.Name)
	fmt.Println("fileLen (WaitFile)", fileLen)

	if err != nil {
		fmt.Println("WaitFile : ", err)
	}

	for i := 0; i < ((fileLen/chunkSize)+1); i++ {
		event_rc = c.Wait("evt", map[string]string{"topic": "fileTransfer", "event": "fileTransferStart"}, 5, iterator)
		payload = event_rc.GetPayload()

		fmt.Println("payload (WaitFile)", payload)

		dataString = strings.Split(payload, ",")

		//Convert payload back to string and Add data to file
		fmt.Println("strings.Split()",strings.Split(strings.Trim(dataString[1], "[]"), " "))

		for _, ps := range strings.Split(strings.Trim(dataString[1], "[]"), " ") {
			pi,_ := strconv.Atoi(ps)
			file.Data = append(file.Data,byte(pi))
		}
	}	
	file.m.Unlock()
	file.Status="ready"
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
