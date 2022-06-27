package shoset

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ditrit/shoset/msg"
)

type FileChunkHandler struct{}

// GetFileChunk :
func (eh *FileChunkHandler) Get(c *ShosetConn) (msg.Message, error) {
	var fileChunkMessage msg.FileChunkMessage
	err := c.ReadMessage(&fileChunkMessage) // ???
	return fileChunkMessage, err
}

// HandleFileChunk :
func (eh *FileChunkHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	fileChunkMessage := message.(msg.FileChunkMessage)
	if state := c.GetCh().Queue["fileChunk"].Push(fileChunkMessage, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
		eh.SendEventConn(c, fileChunkMessage)
	}
	return nil
}

// SendFileChunk :
// FileChunkMessage are sent using ShosetConn not Shoset
func (eh *FileChunkHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("FileChunkHandler.Send not implemented")
}

// SendFileChunk : //A dégager
func (eh *FileChunkHandler) SendEventConn(c *ShosetConn, fileChunkMessage interface{}) {
	_, err := c.WriteString("fileChunk")
	if err != nil {
		log.Warn().Msg("couldn't write string evt : " + err.Error())
		return
	}
	err = c.WriteMessage(fileChunkMessage)
	if err != nil {
		log.Warn().Msg("couldn't write message evt : " + err.Error())
		return
	}
}

// WaitFileChunk :
func (eh *FileChunkHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	term := make(chan *msg.Message, 1)
	cont := true // ??

	go func() {
		for cont {
			//Check message presence in two steps to avoid accessing attributs of <nil>
			cell := replies.Get()
			if cell == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			message := cell.GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			//fmt.Println("(WaitChunk) args[fileName] : ", args["fileName"], "args[firstChunk] : ", args["firstChunk"])
			//fmt.Println("(WaitChunk) message.(msg.FileChunkMessage).GetFileName() : ", message.(msg.FileChunkMessage).GetFileName())

			/* Send back chunk only if :
			- Provided FileName is blank ( and it is the first chunk of the File)
			- FileName match the provided FileName
			- This Filename is not already handled by another WaitFile thread
			*/
			// Vérifier ques les chunk avec le mauvais FileName ne sont pas consommés (lancer 2 transferts en même temps)
			fileName := message.(msg.FileChunkMessage).GetFileName()
			fmt.Println("(WaitChunk) CheckIfFileIsHandled(fileName) : ",msg.CheckIfFileIsHandled(fileName))
			fmt.Println("(WaitChunk) msg.HandledFiles : ",msg.HandledFiles1.HandledFilesList)
			if (args["fileName"] == "" && args["firstChunk"] == "true" && !msg.CheckIfFileIsHandled(fileName)) || (args["fileName"] == fileName) {
				//fmt.Println("Sending chunk")
				term <- &message
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}
