package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type File struct {
	Name   string
	Path   string
	Data   []byte
	Status string //empty, ready, imcomplete, busy
	m      sync.Mutex
}

// Create a new File object from a file on disk.
func NewFile(path string) (*File, error) {
	file := File{}
	file.m.Lock()
	file.Status = "Empty"

	file.Name = filepath.Base(path)
	file.Path = filepath.Dir(path)

	//fmt.Println("(NewFile) file.Path : ",file.Path)
	var err error
	file.Data, err = os.ReadFile(path)

	file.Status = "ready"
	file.m.Unlock()
	return &file, err
}

// Create a new File with empty Name, Path and Data
func NewEmptyFile() *File {
	file := File{}
	file.m.Lock()
	file.Status = "Empty"
	file.m.Unlock()
	return &file
}

// Write file to disk with specidied path (update path with specified path)
func (file *File) WriteToDisk(path string) error {
	file.m.Lock()
	file.Status = "Busy"

	//fmt.Println("File writen to disk (WriteToDisk) : ", filepath.Join(path,file.Name))
	// Ajouter log de l'écriture sur le disque
	file.Path = path
	err := os.WriteFile(filepath.Join(path, file.Name), file.Data, 0222) // Revoir niveau droit

	file.Status = "ready"
	defer file.m.Unlock()
	return err
}

func (file *File) String() string {
	var result string
	result += "Name (file) : " + file.Name + "\n"
	result += "Path : " + file.Path + "\n"
	result += "Data (len) : " + fmt.Sprint(len(file.Data)) + "\n"
	result += "Status : " + file.Status + "\n"

	return result
}
