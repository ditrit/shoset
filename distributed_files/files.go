package files

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

func NewEmptyFile() *File {
	file := File{}
	file.m.Lock()
	file.Status = "Empty"
	file.Name=""
	file.Path=""
	file.Data=[]byte{}	
	
	file.m.Unlock()
	return &file
}

func (file *File) WriteToDisk(path string) error {
	file.m.Lock()
	file.Status = "Busy"
	var err error = nil
	fmt.Println(path+"/"+file.Name)
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

func (files *Files) PrintAllFiles() {
	result := "\nList of imported files :"
	for name, file := range files.FilesMap {
		result+="\nName : "+name+"\n"
		result+="Path : "+file.Path+"\n"
		result+="Data (len) : "+fmt.Sprint(len(file.Data))+"\n"
		result+="Status : "+file.Status+"\n"
	}
	fmt.Println(result)
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
