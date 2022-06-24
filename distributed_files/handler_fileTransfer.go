package files

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ditrit/shoset"
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

// WaitFile :
func (transfer *FileTransfer) WaitFile(c *shoset.Shoset) *File {
	//Créer un nouveua fichier seulement si il n'éxiste pas
	file := NewEmptyFile()

	file.m.Lock()

	data := make(map[int]([]byte)) // Put it directly in FileTransfer

	fileLen := 0

	//Get File info in first message :
	iterator := msg.NewIterator(c.Queue["fileChunk"])
	//c.Wait("fileChunk", map[string]string{}, 5, iterator).(msg.FileChunkMessage)
	for { //
		chunk_rc := c.Handlers["fileChunk"].Wait(c, iterator, map[string]string{}, 2)
		if chunk_rc != nil { //
			chunk_rc := (*chunk_rc).(msg.FileChunkMessage)
			fmt.Println("chunk_rc (WaitFile) : ", chunk_rc)

			//Définir le fichier sur lequel on travail à la reception du premier chunk
			file.Name = chunk_rc.GetFileName()
			fileLen = chunk_rc.GetFileLen()
			//Vérifier que le chunk appartient au bon fichier (tester la réjection)
			if !(file.Name == chunk_rc.GetFileName() && fileLen == chunk_rc.GetFileLen()) {
				continue
			}

			chunkNumber := chunk_rc.GetChunkNumber()

			//Modifier la liste des reçus :
			//vérifier qu'il n'est pas déjà dans la liste :

			if !transfer.chunkAlreadyReceived(chunkNumber) { // (tester la réjection)
				transfer.receivedChunks = append(transfer.receivedChunks, chunkNumber)
				//Mettre la liste par ordre croissant  :
				sort.Ints(transfer.receivedChunks)

				data[chunkNumber] = chunk_rc.GetPayloadByte()
			} else {
				fmt.Println("(WaitFile) Chunk already received : ", chunkNumber)
			}

			//Vérifier que le fichier est complet :
			fmt.Println("(WaitFile) len(transfer.receivedChunks)*chunkSize : ",len(transfer.receivedChunks)*chunkSize,"fileLen : ", fileLen)
			if len(transfer.receivedChunks)*chunkSize >= fileLen {
				fmt.Println("(WaitFile) Fichier complet ! ",file.Name)
				break
			}

		} else {
			break
		}
	}

	// Mettre les chunks dans l'ordre dans le File
	for i := 0; i < int(math.Ceil((float64(fileLen) / float64(chunkSize)))); i++ {
		file.Data = append(file.Data, data[i]...)
	}

	file.m.Unlock()
	file.Status = "ready"
	return file
}

func (transfer *FileTransfer) chunkAlreadyReceived(chunkNumber int) bool {
	for _, a := range transfer.receivedChunks {
		if a == chunkNumber {
			return true
		}
	}
	return false
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
