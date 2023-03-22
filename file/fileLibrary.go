package fileSync

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ditrit/shoset/msg"
)

/*
This file is used to describe a libray linked to a shoset
It contains some tools like manipulating files and directories and loading the library
*/

type FileLibrary interface {
	// Lock the library mutex. A real file is added by the user. We create the copy file and add the syncFile to the library
	UploadFile(file File) (SyncFile, error)
	// create a syncFile from the copy file and the uuid
	// used when we receive a new file from another node
	CreateFile(file File, uuid string) (SyncFile, error)
	GetFile(uuid string) (SyncFile, error)
	GetDir() string
	// Update all the files in the library and create leechers if necessary
	// it compare the list of file state given with our library
	UpdateLibrary(listFiles []FileState, conn ShosetConn)
	// Print info of every files in the library
	PrintAllFiles()
	// get the some basic info about our library in a message format
	GetMessageLibrary() (*msg.FileMessage, error)
	// Add all the files that are sync to the library
	LoadLibrary() error
	SetFileTransfer(fileTransfer FileTransfer)
	// not used yet
	// we want to keep a trace of a file that has been deleted
	DeleteFile(uuid string)
	GetHash() string
	// move a file/folder to another place inside the library
	Move(from string, to string) error
	// add a file to the library
	// the file has already been copied in the library, we just have to add it to the library and to synchronise it
	Add(libraryPath string) error
	// remove a file from the library
	Remove(path string) error
	// a file has been modified in the library path
	Modify(libraryPath string) error
}

type FileLibraryImpl struct {
	FilesMap     map[string]SyncFile // Links a uuid to a pointer to a File object -> []*File
	PathUUIDMap  map[string]string   // Links a path (realPath) to a uuid
	libraryDir   string              //path to the library directory : directory where all the files are stored
	hash         string
	FileTransfer FileTransfer
	m            sync.Mutex
}

// Create new empty FileLibrary
func NewFileLibrary(libraryDir string) (*FileLibraryImpl, error) {
	fileLibrary := FileLibraryImpl{}
	fileLibrary.FilesMap = make(map[string]SyncFile)
	fileLibrary.PathUUIDMap = make(map[string]string)
	fileLibrary.libraryDir = libraryDir

	err := fileLibrary.LoadLibrary()

	return &fileLibrary, err
}

// Add a new file to the FileLibrary
func (fileLibrary *FileLibraryImpl) addNewSyncFile(syncFile SyncFile) {
	fileLibrary.FilesMap[syncFile.GetUUID()] = syncFile
	fileLibrary.PathUUIDMap[syncFile.GetRealFile().GetRelativePath()+syncFile.GetRealFile().GetName()] = syncFile.GetUUID()
	fmt.Println("New file added to the library : ", syncFile)
	fileLibrary.calculateHash()
}

// Lock the library mutex. A real file is added by the user. We create the copy file and add the syncFile to the library
func (fileLibrary *FileLibraryImpl) UploadFile(file File) (SyncFile, error) {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	return fileLibrary.uploadFile(file)
}

// a real file is added by the user. We create the copy file and add the syncFile to the library
func (fileLibrary *FileLibraryImpl) uploadFile(file File) (SyncFile, error) {
	syncFile := NewSyncFile(fileLibrary.libraryDir)
	syncFile.SetFileTransfer(fileLibrary.FileTransfer)
	syncFile.RealFile = file
	var err error
	syncFile.CopyFile, err = LoadFile(fileLibrary.libraryDir, RealToCopyPath(file.GetRelativePath()), file.GetName())
	if err != nil {
		return nil, err
	}
	syncFile.WriteRealToCopy()
	if err != nil {
		return nil, err
	}
	fileLibrary.addNewSyncFile(syncFile)
	return syncFile, nil
}

// create a syncFile from the copy file and the uuid
// used when we receive a new file from another node
func (fileLibrary *FileLibraryImpl) CreateFile(file File, uuid string) (SyncFile, error) {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	syncFile := NewSyncFile(fileLibrary.libraryDir)
	syncFile.SetFileTransfer(fileLibrary.FileTransfer)
	syncFile.SetUUID(uuid)
	syncFile.CopyFile = file
	var err error
	syncFile.RealFile, err = LoadFile(fileLibrary.libraryDir, CopyToRealPath(file.GetRelativePath()), file.GetName())
	if err != nil {
		return nil, err
	}
	fileLibrary.addNewSyncFile(syncFile)
	return syncFile, nil
}

func (fileLibrary *FileLibraryImpl) GetFile(uuid string) (SyncFile, error) {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	file, ok := fileLibrary.FilesMap[uuid]
	var err error
	if !ok {
		err = fmt.Errorf("File %s not found in library", uuid)
	}
	return file, err
}
func (fileLibrary *FileLibraryImpl) GetDir() string {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	return fileLibrary.libraryDir
}

// Update all the files in the library and create leechers if necessary
// it compare the list of file state given with our library
func (fileLibrary *FileLibraryImpl) UpdateLibrary(listFiles []FileState, conn ShosetConn) {
	for _, fileState := range listFiles {
		syncFile, err := fileLibrary.GetFile(fileState.UUID)
		if err != nil {
			go fileLibrary.FileTransfer.AskInfoFile(conn, fileState)
		} else {
			err := syncFile.UpdateFile(fileState, conn)
			if err != nil { // conflict
				// TODO
				fmt.Println(err)
			}
		}

	}
}

// Print info of every files in the library
func (fileLibrary *FileLibraryImpl) PrintAllFiles() {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	result := "\nList of imported files in fileLibrary :"
	for name, file := range fileLibrary.FilesMap {
		result += "\nName (library) : " + name + "\n"
		result += file.String()
	}

	fmt.Println(result)
}

// get the some basic info about our library in a message format
func (fileLibrary *FileLibraryImpl) GetMessageLibrary() (*msg.FileMessage, error) {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	listFileState := make([]msg.FileStateMessage, 0)
	for _, syncFile := range fileLibrary.FilesMap {
		listFileState = append(listFileState, syncFile.GetFileState().FileStateMessage())
	}
	info := msg.FileMessage{MessageName: "sendLibrary",
		Library: listFileState,
	}
	info.InitMessageBase()
	return &info, nil
}

// Add all the files that are sync to the library
func (fileLibrary *FileLibraryImpl) LoadLibrary() error {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	path, err := os.Open(filepath.Join(fileLibrary.libraryDir, PATH_COPY_FILES, ".info"))
	path.Close()
	if os.IsNotExist(err) { // if the repertory doesn't exist
		// we create it
		err = os.MkdirAll(filepath.Join(fileLibrary.libraryDir, PATH_COPY_FILES, ".info"), 0755)
		return err
	} else if err != nil {
		return err
	}
	path, err = os.Open(filepath.Join(fileLibrary.libraryDir, PATH_COPY_FILES, ".info"))
	if err != nil {
		return err
	}
	files, err := path.ReadDir(0)
	if err != nil {
		return err
	}

	list_files := []string{}
	for _, v := range files {
		list_files = append(list_files, v.Name())
	}
	path.Close()

	for _, fileName := range list_files {
		syncFile, err := LoadSyncFileInfo(fileLibrary.libraryDir, filepath.Join(fileLibrary.libraryDir, PATH_COPY_FILES, ".info", fileName))
		if err != nil {
			return err
		}
		syncFile.SetFileTransfer(fileLibrary.FileTransfer)
		fileLibrary.addNewSyncFile(syncFile)
	}
	return nil
}

func (fileLibrary *FileLibraryImpl) SetFileTransfer(fileTransfer FileTransfer) {
	fileLibrary.FileTransfer = fileTransfer
}

// not used yet
// we want to keep a trace of a file that has been deleted
func (fileLibrary *FileLibraryImpl) DeleteFile(uuid string) {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	delete(fileLibrary.FilesMap, uuid)
}

// calculate the hash of the library (with the hash, name, relative path and version of every file)
func (fileLibrary *FileLibraryImpl) calculateHash() {
	toHash := ""
	for _, file := range fileLibrary.FilesMap {
		toHash += file.GetCopyFile().GetHash()
		toHash += file.GetCopyFile().GetName()
		toHash += file.GetCopyFile().GetRelativePath()
		toHash += strconv.Itoa(file.GetCopyFile().GetVersion())
	}
	fileLibrary.hash = Hash([]byte(toHash))
}

func (fileLibrary *FileLibraryImpl) GetHash() string {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	return fileLibrary.hash
}

// move a file/folder to another place inside the library
func (fileLibrary *FileLibraryImpl) Move(from string, to string) error {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	op := Operation{Name: "modify", File: from, NewFile: to}
	fromFile, err := os.Stat(filepath.Join(fileLibrary.libraryDir, from))
	if err != nil {
		return fmt.Errorf("can't get the file : %s", err)
	}
	if !fromFile.IsDir() { // if we are moving a file
		relpath := filepath.Dir(from)
		uuid, ok := fileLibrary.PathUUIDMap[relpath]
		if ok { // if we have the file in the library
			syncFile, ok := fileLibrary.FilesMap[uuid]
			if ok {
				err := syncFile.ApplyOperationFromMe(op)
				if err != nil {
					return err
				}
			} else {
				delete(fileLibrary.PathUUIDMap, relpath)
			}
		} else { // if we don't have the file in the library
			return fmt.Errorf("the file is not in the library")
		}
	} else { // if it's a directory
		// we move all the files inside
		for _, syncFile := range fileLibrary.FilesMap {
			if strings.HasPrefix(syncFile.GetRealFile().GetRelativePath(), from) {
				err := syncFile.ApplyOperationFromMe(op)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// add a file to the library
// the file has already been copied in the library, we just have to add it to the library and to synchronise it
func (fileLibrary *FileLibraryImpl) Add(libraryPath string) error {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	// we upload it
	relPath := PathToRelativePath(libraryPath, fileLibrary.libraryDir)
	relativePath := Dir(relPath)
	fileName := filepath.Base(libraryPath)

	_, ok := fileLibrary.PathUUIDMap[relPath]
	if ok {
		return fmt.Errorf("the file is already in the library : %s", relativePath)
	}

	file, err := LoadFile(fileLibrary.libraryDir, relativePath, fileName)
	if err != nil {
		return fmt.Errorf("can't add the file : %s", err)
	}
	syncFile, err := fileLibrary.uploadFile(file)
	if err != nil {
		return fmt.Errorf("can't upload the file : %s", err)
	}
	err = syncFile.ApplyOperationFromMe(Operation{Name: "create"})
	return err
}

// remove a file from the library
func (fileLibrary *FileLibraryImpl) Remove(path string) error {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	relPath := PathToRelativePath(path, fileLibrary.libraryDir)
	uuid, ok := fileLibrary.PathUUIDMap[relPath]
	if ok { // if we have the file in the library
		file, ok := fileLibrary.FilesMap[uuid]
		if ok { // if we don't have the file in the library
			err := file.ApplyOperationFromMe(Operation{Name: "remove"})
			if err != nil {
				return err
			}
		} else {
			delete(fileLibrary.PathUUIDMap, relPath)
		}
	} else { // if we don't have the file in the library
		// maybe it was a directory
		for _, syncFile := range fileLibrary.FilesMap {
			if strings.HasPrefix(syncFile.GetRealFile().GetRelativePath(), relPath) {
				err := syncFile.ApplyOperationFromMe(Operation{Name: "remove"})
				if err != nil {
					return err
				}
			}
		}
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("can't remove the file : %s", err)
		}
	}
	return nil
}

// a file has been modified in the library path
func (fileLibrary *FileLibraryImpl) Modify(libraryPath string) error {
	fileLibrary.m.Lock()
	defer fileLibrary.m.Unlock()
	relPath := PathToRelativePath(libraryPath, fileLibrary.libraryDir)
	uuid, ok := fileLibrary.PathUUIDMap[relPath]
	if ok { // if we have the file in the library
		syncFile, ok := fileLibrary.FilesMap[uuid]
		if ok {
			if syncFile.GetStatus() == "downloading" {
				return fmt.Errorf("the file %s is currently downloading", libraryPath)
			}
			err := syncFile.ApplyOperationFromMe(Operation{Name: "modify"})
			if err != nil {
				return fmt.Errorf("can't modify the file : %s", err)
			}
		} else {
			delete(fileLibrary.PathUUIDMap, relPath)
		}
	} else { // we don't have this file in the library
		return fmt.Errorf("the file is not in the library")
	}
	return nil
}
