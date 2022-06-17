package files

import (
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	name string
	// path string ??????
	data   []byte
	status string
}

func NewFile(path string) (File, error) {
	file := File{}
	file.status = "Empty"
	var err error

	file.name = filepath.Base(path)

	file.data, err = os.ReadFile(path)
	file.status = "ready"
	return file, err
}

func (file *File) WriteToDisk(path string) error {
	file.status = "Busy"
	var err error = nil
	//fmt.Println(path + file.name)
	err = os.WriteFile(path+file.name, file.data, 0222)
	file.status = "ready"
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

	file := File{}
	file.status = "Empty"
	var err_read error

	file, err_read = NewFile(path)

	if err_read != nil {
		fmt.Println(err_read)
	}

	file.status = "ready"

	files.FilesMap[file.name] = &file
}

func (files *Files) GetAllFiles() {
	for _, i := range files.FilesMap {
		fmt.Println(i.name)
	}
}

type FileTranfer struct {
	transferType   string   // "tx" or "rx" (Lock the data of the file for the duration of transfer)
	file           *File    //File to be transfered
	receivedChunks []int    //List of the ids of chunks received
	sources        []string //[]*Shoset.Conn List of connexions involved in the transfer
	/*
		map[*Shoset.Conn] ([]int) List of chunks requested by a connexion
		Requested chunks must also be in received or the file is complete
	*/
	requestedChunks map[string][]int
}

func (file *File) NewFileTransfer(transferType string) FileTranfer {
	var transfer FileTranfer
	transfer.transferType = "tx"
	transfer.file = file
	transfer.receivedChunks = []int{}
	transfer.sources = []string{}                     //[]*Shoset.Conn
	transfer.requestedChunks = make(map[string][]int) // map[*Shoset.Conn] ([]int)
	return transfer
}