package file

import (
	"fmt"
	"math"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

//Implementer handler interface

// Get ?? :

// HandleFileTranfer :
func (fileTransfer *FileTransfer) HandleTransfer() {
	fileChunk := []byte{}

	lenFile := len(fileTransfer.file.Data)
	//Send requested chunks :
	for conn, chRqByConn := range fileTransfer.requestedChunks {		
		message := msg.NewFileChunkMessage(fileTransfer.file.Name, lenFile, -1, nil)
		fmt.Println("message (HandleTransfer) : ", message) //
		err := conn.SendMessage(message)
		if err != nil {
			fmt.Println("sendChunk : ", err)
		}

		time.Sleep(50 * time.Millisecond)
		for _, chunk := range chRqByConn {
			//fmt.Println("chunk (HandleTransfer) : ", chunk)
			//Create data chunk :

			if (chunk+1)*chunkSize < lenFile {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : (chunk+1)*chunkSize]
			} else {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : lenFile]
			}

			//Generate message :
			message := msg.NewFileChunkMessage(fileTransfer.file.Name, lenFile, chunk, fileChunk)
			fmt.Println("message (HandleTransfer) : ", message) //
			err := conn.SendMessage(message)

			time.Sleep(50 * time.Millisecond) //Necessary to make it work

			if err != nil {
				fmt.Println("sendChunk : ", err)
			}
		}
	}
	//Check if no other chunks are requested :
}

// WaitFile :
// Receive chunks and reassemble File from chunks
func (transfer *FileTransfer) WaitFile(iterator *msg.Iterator) *File {
	transfer.file.m.Lock()

	if iterator == nil {
		iterator = msg.NewIterator(transfer.shosetCom.Queue["fileChunk"])
	}

	var fileName string

	chunk_rc := transfer.shosetCom.Handlers["fileChunk"].Wait(transfer.shosetCom, iterator, map[string]string{"fileName": fileName}, 2)
	if chunk_rc != nil {
		chunk_rc := (*chunk_rc).(msg.FileChunkMessage)
		transfer.file.Name = chunk_rc.GetFileName()
		fmt.Println("chunk_rc (WaitFile) : ", chunk_rc)
	}
	transfer.waitFile(iterator)
	defer transfer.file.m.Unlock()
	return transfer.file
}

// WaitFile :
// Receive chunks and reassemble File from chunks
func (transfer *FileTransfer) WaitFileName(iterator *msg.Iterator, fileName string) *File {
	transfer.file.m.Lock()

	if iterator == nil {
		iterator = msg.NewIterator(transfer.shosetCom.Queue["fileChunk"])
	}

	chunk_rc := transfer.shosetCom.Handlers["fileChunk"].Wait(transfer.shosetCom, iterator, map[string]string{"fileName": fileName}, 2)
	if chunk_rc != nil { //
		chunk_rc := (*chunk_rc).(msg.FileChunkMessage)
		transfer.file.Name = chunk_rc.GetFileName()
		fmt.Println("chunk_rc (WaitFile) : ", chunk_rc)
	}
	transfer.waitFile(iterator)
	defer transfer.file.m.Unlock()
	return transfer.file
}

// WaitFile :
// Receive chunks and reassemble File from chunks
// Internal method handling the actual file reception
func (transfer *FileTransfer) waitFile(iterator *msg.Iterator) {

	var fileLen int // Need to to be accessible outside of for loop

	data := make(map[int]([]byte)) // Put it directly in FileTransfer ?
	for {
		chunk_rc := transfer.shosetCom.Handlers["fileChunk"].Wait(transfer.shosetCom, iterator, map[string]string{"fileName": transfer.file.Name}, 2)
		if chunk_rc != nil {
			chunk_rc := (*chunk_rc).(msg.FileChunkMessage)
			fmt.Println("chunk_rc (WaitFile) : ", chunk_rc)

			fileLen = chunk_rc.GetFileLen()
			chunkNumber := chunk_rc.GetChunkNumber()

			//vérifier qu'il n'est pas déjà dans la liste :
			if !transfer.chunkAlreadyReceived(chunkNumber) { // (tester la réjection)
				//Add chunkNumber to the list of received chunks
				transfer.receivedChunks = append(transfer.receivedChunks, chunkNumber)

				//Store received data
				data[chunkNumber] = chunk_rc.GetPayloadByte()
			} else {
				fmt.Println("(WaitFile) Chunk already received : ", chunkNumber)
			}
			// Stop reception if every chunk was received
			if len(transfer.receivedChunks)*chunkSize >= fileLen {
				fmt.Println("(WaitFile) Fichier complet ! ", transfer.file.Name)
				break
			}
		} else {
			break
		}
	}
	// Reconstruct File.Data from received chunks
	for i := 0; i < int(math.Ceil((float64(fileLen) / float64(chunkSize)))); i++ {
		transfer.file.Data = append(transfer.file.Data, data[i]...)
	}
	transfer.file.Status = "ready"
}

// Check if a chunk number was already received
func (transfer *FileTransfer) chunkAlreadyReceived(chunkNumber int) bool {
	for _, a := range transfer.receivedChunks {
		if a == chunkNumber {
			return true
		}
	}
	return false
}


//Request to be sent a file
// Revoyer un transfer ou le fichier
func RequestFile(c *shoset.Shoset, name string, originAdress string) *File {
	//Create Transfer :
	transfer := NewFileTransferRx(c, originAdress)
	
	//Request File infos :
	message := msg.NewFileChunkMessage(name, -1, -2, nil)
	fmt.Println("message (HandleTransfer) : ", message) //
	
	// Request file to every conn in expectedChunks
	for conn,_ :=range transfer.expectedChunks {
		err := conn.SendMessage(message)
		if err != nil {
			fmt.Println("sendChunk : ", err)
		}
	}

	//Receive File :
	 return transfer.WaitFileName(nil,name)
}

func HandleRequestedFile(){
	
}