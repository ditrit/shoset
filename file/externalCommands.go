package fileSync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ditrit/shoset/msg"
	guuid "github.com/kjk/betterguid"
)

/*
This package provide accessible commands to the user.
The commands are : move, delete, add, modify
Use these commands to modify a file in the library to do it safely.
It asks to the other node for their permission to modify the library.
*/

type ExternalCommands struct {
	FileTransfer FileTransfer
	locked       bool
	currentUUID  string
	addressMap   map[string]bool
	nbConn       int
	answerChan   chan bool

	m sync.Mutex
}

func NewExternalCommands(fileTransfer FileTransfer) *ExternalCommands {
	ec := new(ExternalCommands)
	ec.FileTransfer = fileTransfer
	ec.addressMap = make(map[string]bool)
	ec.answerChan = make(chan bool, 1)

	return ec
}

// rename or move a file / folder inside the library
func (ec *ExternalCommands) Move(previousPath string, newPath string) error {
	err := ec.VerifyPath(previousPath)
	if err != nil {
		return err
	}
	err = ec.VerifyPath(newPath)
	if err != nil {
		return err
	}
	err = ec.Lock()
	defer ec.Unlock()
	if err != nil {
		return err
	}

	err = ec.FileTransfer.GetLibrary().Move(previousPath, newPath)
	return err
}

// delete a file / folder from the library
func (ec *ExternalCommands) Delete(path string) error {
	err := ec.VerifyPath(path)
	if err != nil {
		return err
	}
	err = ec.Lock()
	defer ec.Unlock()
	if err != nil {
		return err
	}

	err = ec.FileTransfer.GetLibrary().Remove(path)
	return err
}

// add a new file to the library
// path : full path of the file to add
// libraryPath : path of the file in the library
// if libraryPath does not exist, it will be created
func (ec *ExternalCommands) Add(path string, libraryPath string) error {
	fmt.Println("add : ", path, " to ", libraryPath, " in the library")
	file, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("can't add the file : %s", err)
	}
	if file.IsDir() {
		return fmt.Errorf("can't add a directory")
	}
	_, err = os.Stat(libraryPath)
	if err == nil {
		return fmt.Errorf(libraryPath + " already exist in the library")
	}
	if !IsInDir(libraryPath, ec.FileTransfer.GetLibrary().GetDir()) {
		return fmt.Errorf(path + " is not a path in the library")
	}

	err = ec.Lock()
	defer ec.Unlock()
	if err != nil {
		return err
	}
	// create the libraryPath if it does not exist
	directory := filepath.Dir(libraryPath)
	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	libraryPathFile, err := os.Create(libraryPath)
	if err != nil {
		return err
	}
	pathFile, err := os.Open(path)
	if err != nil {

		return err
	}
	_, err = io.Copy(libraryPathFile, pathFile)
	if err != nil {
		return err
	}
	err = ec.FileTransfer.GetLibrary().Add(libraryPath)
	return err
}

// modify a file in the library
// it replace the file content in the library by the file in the path
func (ec *ExternalCommands) Modify(path string, libraryPath string) error {
	pathInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if pathInfo.IsDir() {
		return fmt.Errorf(path + " is a directory, you can't 'modify' a directory")
	}
	err = ec.VerifyPath(libraryPath)
	if err != nil {
		return err
	}
	err = ec.Lock()
	defer ec.Unlock()
	if err != nil {
		return err
	}

	err = ec.FileTransfer.GetLibrary().Modify(libraryPath)
	return err
}

func (ec *ExternalCommands) Lock() error {
	ec.m.Lock()
	ec.currentUUID = guuid.New()
	askLocked := msg.FileMessage{
		MessageName: "askLibraryLocked",
		FileUUID:    ec.currentUUID,
	}
	askLocked.InitMessageBase()
	ec.FileTransfer.Broadcast(&askLocked)
	ec.m.Unlock()
	err := ec.WaitAnswer()
	return err
}

func (ec *ExternalCommands) WaitAnswer() error {
	select {
	case answer := <-ec.answerChan:
		ec.m.Lock()
		defer ec.m.Unlock()
		ec.locked = answer
		if !answer {
			err := fmt.Errorf("library is locked because : some nodes have already locked their library : %T", ec.addressMap)
			ec.addressMap = make(map[string]bool)
			return err
		} else {
			ec.addressMap = make(map[string]bool)
			return nil
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("library is locked because : timeout")
	}
}

func (ec *ExternalCommands) Unlock() {
	ec.m.Lock()
	defer ec.m.Unlock()
	ec.locked = false
}

func (ec *ExternalCommands) VerifyPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !IsInDir(path, ec.FileTransfer.GetLibrary().GetDir()) {
		return fmt.Errorf(path + " is not a path in the library")
	}
	return nil
}

func (ec *ExternalCommands) ReceiveAnswer(answer *msg.FileMessage, address string, nbConn int) {
	if answer.FileUUID == ec.currentUUID {
		ec.m.Lock()
		ec.nbConn = nbConn
		ec.addressMap[address] = answer.AnswerLocked
		if len(ec.addressMap) == ec.nbConn {
			//fmt.Println("we have all the answers for locking: ", ec.addressMap)
			answer := true
			for _, v := range ec.addressMap {
				if v { // if someone answered that it was locked
					answer = false
					ec.answerChan <- false
					break
				}
			}
			if answer {
				ec.answerChan <- true
			}
		}
		ec.m.Unlock()
	}
}

func (ec *ExternalCommands) IsLocked(ask *msg.FileMessage) msg.FileMessage {
	ec.m.Lock()
	defer ec.m.Unlock()
	answer := msg.FileMessage{
		MessageName:  "answerLibraryLocked",
		AnswerLocked: ec.locked,
		FileUUID:     ask.FileUUID,
	}
	answer.InitMessageBase()
	return answer
}
