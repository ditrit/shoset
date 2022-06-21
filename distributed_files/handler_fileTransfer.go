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

//func (c *ShosetConn) SendMessage(msg msg.Message)

func (fileTransfer *FileTranfer) HandleTransfer() {
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
		fileChunk = fmt.Sprint(chunkNumber) + "," + fmt.Sprint(file.Data[chunkNumber*chunkSize:len(file.Data)-1])
	}

	event := msg.NewEventClassic("fileTransfer", "fileTransferStart", fileChunk)

	fmt.Println("event (sendChunk)", event)

	//Send event :
	time.Sleep(100 * time.Millisecond)
	error := destination.SendMessage(event)

	if error != nil {
		fmt.Println("sendChunk : ",error)
	}
}

//

// WaitChunk :
func (file *File) WaitChunk() {
	//Find chunk number :

	//Add to file data :
}
