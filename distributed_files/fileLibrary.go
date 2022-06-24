package file

import (
	"fmt"
)

type FileLibrary struct {
	FilesMap map[string]*File //Links a name to a pointer to a File object
}

// Create new empty FileLibrary
func NewFiles() FileLibrary {
	var fileLibrary FileLibrary
	fileLibrary.FilesMap = make(map[string]*File)

	return fileLibrary
}

// Add a new file tot the FileLibrary
func (fileLibrary *FileLibrary) AddNewFile(path string) {

	file := &File{}
	file.Status = "Empty"
	var err_read error

	file, err_read = NewFile(path)

	if err_read != nil {
		fmt.Println(err_read)
	}

	file.Status = "ready"

	fileLibrary.FilesMap[file.Name] = file
}

//Print info of every files in the library
func (fileLibrary *FileLibrary) PrintAllFiles() {
	result := "\nList of imported files in fileLibrary :"
	for name, file := range fileLibrary.FilesMap {
		result += "\nName (library) : " + name + "\n"
		result += file.String()
	}
	fmt.Println(result)
}

//Write every file in the library to disk in the specified folder
func (fileLibrary *FileLibrary) WriteAllToDisk(path string) error {
	var err error
	for _, s := range fileLibrary.FilesMap {
		err = s.WriteToDisk(path)
		if err != nil {
			return err
		}
	}
	return err
}
