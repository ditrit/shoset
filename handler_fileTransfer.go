package shoset

import (
	"errors"
	"fmt"
	"time"

	fileMod "github.com/ditrit/shoset/file"
	"github.com/ditrit/shoset/msg"
)

// EventHandler implements MessageHandlers interface.
type FileHandler struct {
	fileMod.FileTransferImpl
}

// Get returns the message for a given ShosetConn.
func (fh *FileHandler) Get(c *ShosetConn) (msg.Message, error) {
	var fm msg.FileMessage
	err := c.GetReader().ReadMessage(&fm)
	if err != nil {
		return nil, err
	}
	if contains([]string{"authorised", "unauthorised", "congestion", "askInfo", "interested", "notInterested", "have", "askBitfield", "sendBitfield", "sendInfo", "sendChunk", "askChunk", "sendLibrary", "sendFileOperation", "askLibraryLocked", "answerLibraryLocked"}, fm.MessageName) {
		return fm, err
	}
	return fm, errors.New("message not found : " + fm.MessageName)
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (fh *FileHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	if fh.GetLibrary() == nil { // the library has not been initialised
		c.GetShoset().Logger.Warn().Msg("library not initialised")
		return nil
	}
	fileMessage := message.(msg.FileMessage)
	switch fileMessage.MessageName {
	case "sendChunk":
		c.Logger.Debug().Msg(c.GetLocalAddress() + " received " + fileMessage.MessageName + " from " + c.GetRemoteAddress() + " /" + " begin: " + fmt.Sprint(fileMessage.Begin) + " length: " + fmt.Sprint(fileMessage.Length))
	case "askChunk":
		c.Logger.Debug().Msg(c.GetLocalAddress() + " received " + fileMessage.MessageName + " from " + c.GetRemoteAddress() + " /" + " begin: " + fmt.Sprint(fileMessage.Begin) + " length: " + fmt.Sprint(fileMessage.Length))
	case "sendBitfield":
		c.Logger.Debug().Msg(c.GetLocalAddress() + " received " + fileMessage.MessageName + " from " + c.GetRemoteAddress() + " /" + " with " + fmt.Sprint(fileMessage.Bitfield))
	case "askBitfield":
		c.Logger.Debug().Msg(c.GetLocalAddress() + " received " + fileMessage.MessageName + " from " + c.GetRemoteAddress())

	default:
		c.Logger.Debug().Msg(c.GetLocalAddress() + " received " + fileMessage.MessageName + " from " + c.GetRemoteAddress())
	}
	fh.ReceiveMessage(&fileMessage, c)
	return nil
}

// Send a message
func (fh *FileHandler) Send(s *Shoset, message msg.Message) {
	s.ConnsByLname.Iterate(
		func(lname string, ipAddress string, conn interface{}) {
			if lname == s.GetLogicalName() { // if same Lname as the me
				conn.(*ShosetConn).GetWriter().SendMessage(message)
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (fh *FileHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()
	// Checks every message in the queue before waiting for a new message
	// Checks message presence in two steps to avoid accessing attributs of <nil>
	for {
		cell := replies.Get()
		if cell != nil {
			message := cell.GetMessage()
			if message != nil {
				fileMsg := message.(msg.FileMessage)
				if args["fileName"] == VOID || fileMsg.FileName == args["fileName"] {
					return &message
				}
			}
		} else {
			// Locking Queue to avoid missing a message while preparing the channel to receive events.
			replies.GetQueue().LockQueue()
			break
		}
	}
	// Creating channel
	chNewMessage := make(chan interface{})
	s.MessageEventBus.Subscribe("file", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("file", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait file (timeout).")
			return nil
		case <-chNewMessage:
			//Checks message presence in two steps to avoid accessing fields of <nil>
			s.Logger.Debug().Msg("New message received in Wait file.")
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			fileMsg := message.(msg.FileMessage)
			if args["fileName"] == VOID || fileMsg.FileName == args["fileName"] {
				return &message
			}
		}
	}
}
