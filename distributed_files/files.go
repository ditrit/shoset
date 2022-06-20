package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ditrit/shoset"
)

type File struct {
	Name   string
	Path   string
	Data   []byte
	Status string
	m      sync.Mutex
}

func NewFile(path string) (*File, error) {
	file := File{}
	file.m.Lock()
	file.Status = "Empty"
	var err error

	file.Name = filepath.Base(path)

	file.Data, err = os.ReadFile(path)
	file.Status = "ready"
	file.m.Unlock()
	return &file, err
}

func (file *File) WriteToDisk(path string) error {
	file.m.Lock()
	file.Status = "Busy"
	var err error = nil
	fmt.Println(path + file.Name)
	fmt.Println(string(file.Data))
	file.Path = path
	err = os.WriteFile(path+"/"+file.Name, file.Data, 0222)

	file.Status = "ready"
	defer file.m.Unlock()
	return err
}

type Files struct {
	FilesMap map[string]*File //Links a name to a pointer to a File
}

func NewFiles() Files {
	var files Files
	files.FilesMap = make(map[string]*File)

	return files
}

func (files *Files) AddNewFile(path string) {

	file := &File{}
	file.Status = "Empty"
	var err_read error

	file, err_read = NewFile(path)

	if err_read != nil {
		fmt.Println(err_read)
	}

	file.Status = "ready"

	files.FilesMap[file.Name] = file
}

func (files *Files) GetAllFiles() {
	for _, i := range files.FilesMap {
		fmt.Println(i.Name)
	}
}

func (files *Files) WriteAllToDisk(path string) error {
	var err error
	for _, s := range files.FilesMap {
		err = s.WriteToDisk(path)
		if err != nil {
			return err
		}
	}
	return err
}

type FileTranfer struct {
	sender         *shoset.Shoset
	transferType   string               // "tx" or "rx" (Lock the data of the file for the duration of transfer)
	file           *File                //File to be transfered
	receivedChunks []int                //List of the ids of chunks received
	sources        []*shoset.ShosetConn //List of connexions involved in the transfer
	/*
		List of chunks requested by a connexion
		Requested chunks must also be in received or the file is complete
	*/
	requestedChunks map[*shoset.ShosetConn][]int
}

//destination : adrress (IP:port) ? of the destination
func (file *File) NewFileTransfer(sender *shoset.Shoset, destinationAdress string) FileTranfer {
	var transfer FileTranfer
	transfer.sender = sender
	transfer.transferType = "tx"
	transfer.file = file
	transfer.receivedChunks = []int{}
	transfer.sources = []*shoset.ShosetConn{}                     //[]*Shoset.Conn
	transfer.requestedChunks = make(map[*shoset.ShosetConn][]int) // map[*Shoset.Conn] ([]int)

	//Finding the adress in the established cons of the sender
	var conn *shoset.ShosetConn

	for _, i := range sender.GetConnsByTypeArray("cl") {
		fmt.Println("i.GetRemoteAddress()", i.GetRemoteAddress())
		if i.GetRemoteAddress() == destinationAdress {
			conn = i
		}
	}
	transfer.requestedChunks[conn] = []int{}

	return transfer
}

func (transfer *FileTranfer) String() string {
	result := "\nFileTranfer of " + transfer.file.Name + " :\n"
	result += "sender : " + transfer.sender.String() + "\n"
	result += "transferType : " + transfer.transferType + "\n"
	result += "Amount received : " + fmt.Sprint((len(transfer.receivedChunks))) + "\n"
	result += "Sources (adresses) : "
	for _, i := range transfer.sources {
		result += i.GetRemoteAddress() + ", "
	}
	result += "\n"
	result += "Requested ((adresses) : (amount)) : "
	for conn, chunks := range transfer.requestedChunks {
		result += conn.GetRemoteAddress() + " : " + fmt.Sprint((len(chunks)))+ ", "
	}
	result += "\n"
	return result
}
