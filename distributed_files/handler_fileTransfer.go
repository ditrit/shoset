package file

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ditrit/shoset/msg"
)

//Implementer handler interface

// Get ?? :

// HandleFileTranfer :
func (fileTransfer *FileTransfer) HandleTransfer() {
	fileChunk := []byte{}

	//Send requested chunks :
	for conn, chRqByConn := range fileTransfer.requestedChunks {
		for _, chunk := range chRqByConn {
			fmt.Println("chunk (HandleTransfer) : ", chunk)
			//Create data chunk :
			lenFile := len(fileTransfer.file.Data)
			if (chunk+1)*chunkSize < lenFile {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : (chunk+1)*chunkSize]
			} else {
				fileChunk = fileTransfer.file.Data[chunk*chunkSize : lenFile]
			}

			//Generate message :
			message := msg.NewFileChunkMessage(fileTransfer.file.Name, lenFile, chunk, fileChunk)
			fmt.Println("message (HandleTransfer) : ", message) //
			err := conn.SendMessage(message)

			time.Sleep(10 * time.Millisecond) //Necessary to make it work

			if err != nil {
				fmt.Println("sendChunk : ", err)
			}
		}
	}
	//Check if no other chunks are requested :
}

//Request to be sent a file
func requestFile(name string) {
	//Request File infos :

	//Create File :

	//Request chunks ?? :

	//Start reception :
}

// WaitFile :
// Receive chunks and reassemble File from chunks
func (transfer *FileTransfer) WaitFile() *File {
	transfer.file.m.Lock()

	data := make(map[int]([]byte)) // Put it directly in FileTransfer ?

	var fileLen int
	iterator := msg.NewIterator(transfer.shosetCom.Queue["fileChunk"])
	//transfer.shosetCom.Wait("fileChunk", map[string]string{}, 5, iterator).(msg.FileChunkMessage)
	for { //
		chunk_rc := transfer.shosetCom.Handlers["fileChunk"].Wait(transfer.shosetCom, iterator, map[string]string{}, 2)
		if chunk_rc != nil { //
			chunk_rc := (*chunk_rc).(msg.FileChunkMessage)
			fmt.Println("chunk_rc (WaitFile) : ", chunk_rc)

			//Définir le fichier sur lequel on travail à la reception du premier chunk
			transfer.file.Name = chunk_rc.GetFileName()
			fileLen = chunk_rc.GetFileLen()

			//Vérifier que le chunk appartient au bon fichier (tester la réjection)
			if !(transfer.file.Name == chunk_rc.GetFileName() && fileLen == chunk_rc.GetFileLen()) {
				continue
			}

			chunkNumber := chunk_rc.GetChunkNumber()

			//vérifier qu'il n'est pas déjà dans la liste :
			if !transfer.chunkAlreadyReceived(chunkNumber) { // (tester la réjection)
				//Modifier la liste des reçus :
				transfer.receivedChunks = append(transfer.receivedChunks, chunkNumber)
				//Mettre la liste par ordre croissant  :
				sort.Ints(transfer.receivedChunks)

				//Store received data
				data[chunkNumber] = chunk_rc.GetPayloadByte()
			} else {
				fmt.Println("(WaitFile) Chunk already received : ", chunkNumber)
			}

			//Vérifier que le fichier est complet :
			fmt.Println("(WaitFile) len(transfer.receivedChunks)*chunkSize : ", len(transfer.receivedChunks)*chunkSize, "fileLen : ", fileLen)
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

	transfer.file.m.Unlock()
	transfer.file.Status = "ready"
	return transfer.file
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
