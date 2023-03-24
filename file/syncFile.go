package fileSync

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/ditrit/shoset/msg"
	guuid "github.com/kjk/betterguid"
)

/*
It makes the link between the real file (the file in the library) and the copy file (the file in the .copy folder).
Each SyncFile has a uuid (unique identifier).
This uuid is used to identify the file we are talking about.
It is usefull when a file is renamed.
SyncFile is also used to keep track of the operations made on the file and to update the files in both the library or the library copy.
*/

var opMapScore = map[string]int{"remove": 0, "move": 1, "modify": 2, "create": 3}

type FileState struct {
	UUID          string
	Name          string
	Hash          string
	HashMap       map[int]string
	Version       int
	Path          string
	LastOperation Operation
}

// convert a Filstate structure to a msg.FileStateMessage
func (fs FileState) FileStateMessage() msg.FileStateMessage {
	return msg.FileStateMessage{
		UUID:          fs.UUID,
		Name:          fs.Name,
		Hash:          fs.Hash,
		Version:       fs.Version,
		Path:          fs.Path,
		LastOperation: fs.LastOperation.OperationMessage(),
	}
}

// convert a msg.FileStateMessage structure to a Filstate
func ToFileState(message msg.FileStateMessage) FileState {
	return FileState{
		UUID:          message.UUID,
		Name:          message.Name,
		Hash:          message.Hash,
		Version:       message.Version,
		Path:          message.Path,
		LastOperation: ToOperation(message.LastOperation),
	}
}

type SyncFile interface {
	String() string
	EndDownload() error
	WriteCopyToReal() error
	WriteRealToCopy() error
	// when there is a change on the real file, do it to the Copy file
	ApplyOperationFromMe(op Operation) error
	UpdateFile(fileState FileState, conn ShosetConn) error
	SaveFileInfo() error
	GetRealFile() File
	GetCopyFile() File
	GetUUID() string
	SetUUID(uuid string)
	SetFileTransfer(ft FileTransfer)
	GetLastOperation() Operation
	GetShortInfoMsg() msg.FileMessage
	// like shortInfo but we add the hashmap
	GetFullInfoMsg() msg.FileMessage
	GetFileState() FileState
	SetStatus(status string)
	GetStatus() string
}

type SyncFileImpl struct {
	RealFile     File   // file in the library
	CopyFile     File   // copy of the file in the library (it is in the .copy folder)
	baseFilesDir string // path to the library

	uuid          string       // unique identifier of the file
	lastOperation Operation    // last Operation done on the file (usefull to resolve conflicts if it occurs)
	FileTransfer  FileTransfer // FileTransfer object to send the file to other nodes

	/*
		Different status :
			downloading : the File data is being downloaded
			full : Data is loaded
	*/
	status string //empty, full
	synced bool   // true if the file is synced with the other nodes
	m      sync.RWMutex
}

func NewSyncFile(baseFilesDir string) *SyncFileImpl {
	return &SyncFileImpl{
		baseFilesDir: baseFilesDir,
		status:       "empty",
		synced:       false,
		uuid:         guuid.New(),
	}
}

func (syncFile *SyncFileImpl) String() string {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	var result string
	result += "Name : " + syncFile.RealFile.GetName() + "\n"
	result += "UUID : " + syncFile.uuid + "\n"
	result += "Path : " + filepath.Join(syncFile.baseFilesDir, syncFile.RealFile.GetRelativePath(), syncFile.RealFile.GetName()) + "\n"
	result += "Size: " + fmt.Sprint(syncFile.RealFile.GetSize()) + "\n"
	result += "status : " + syncFile.status + "\n"

	return result
}

func (syncFile *SyncFileImpl) EndDownload() error {
	syncFile.SaveFileInfo()
	syncFile.SetStatus("full")
	syncFile.CopyFile.CloseFile()
	newHash, err := syncFile.CopyFile.CalculateHash()
	if err != nil {
		return err
	}
	if newHash != syncFile.CopyFile.GetHash() {
		hashMap := syncFile.CopyFile.GetHashMap()
		newHashMap, err := syncFile.CopyFile.CalculateHashMap()
		if err != nil {
			return err
		}
		idList := make([]int, 0)
		for i := 0; i < len(hashMap); i++ {
			if hashMap[i] != newHashMap[i] {
				idList = append(idList, i)
			}
		}
		return fmt.Errorf("hash of the file "+syncFile.CopyFile.GetName()+" is not correct because of piece(s)", idList)
	}
	return err
}

func (syncFile *SyncFileImpl) WriteCopyToReal() error {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	syncFile.RealFile.CloseFile()
	syncFile.CopyFile.CloseFile()
	CopyFile, err := syncFile.CopyFile.OpenFile()
	if err != nil {
		return err
	}
	realFile, err := syncFile.RealFile.OpenFile()
	if err != nil {
		return err
	}
	_, err = io.Copy(realFile, CopyFile)
	if err != nil {
		return err
	}
	syncFile.RealFile.CloseFile()
	syncFile.CopyFile.CloseFile()
	err = syncFile.RealFile.UpdateMetadata()
	return err
}

func (syncFile *SyncFileImpl) WriteRealToCopy() error {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	syncFile.RealFile.CloseFile()
	syncFile.CopyFile.CloseFile()
	CopyFile, err := syncFile.CopyFile.OpenFile()
	if err != nil {
		return err
	}
	realFile, err := syncFile.RealFile.OpenFile()
	if err != nil {
		return err
	}
	_, err = io.Copy(CopyFile, realFile)
	if err != nil {
		return err
	}
	syncFile.RealFile.CloseFile()
	syncFile.CopyFile.CloseFile()
	err = syncFile.CopyFile.UpdateMetadata()
	return err
}

// when there is a change on the real file, do it to the Copy file
func (syncFile *SyncFileImpl) ApplyOperationFromMe(op Operation) error {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	if op.NewFile != "" {
		syncFile.RealFile.SetRelativePath(Dir(op.NewFile))
		syncFile.RealFile.SetName(filepath.Base(op.NewFile))
		syncFile.RealFile.SetVersion(syncFile.RealFile.GetVersion() + 1)
	}

	if syncFile.status == "downloading" {
		return fmt.Errorf("the file %s is being downloaded", syncFile.RealFile.GetName())
	}

	// the operation is normally valid and can occur. We apply it
	if op.Name == "create" {
		syncFile.lastOperation = op
		create := syncFile.getShortInfoMsg()
		syncFile.FileTransfer.Broadcast(&create)
	} else if op.Name == "remove" {
		// we remove the leechers for the file (if some were added meanwhile) and the file from the library : it no longer exists
		//syncFile.FileTransfer.Library.DeleteFile(syncFile.CopyFile.GetName())
		syncFile.FileTransfer.DeleteLeecher(syncFile.GetUUID())
	} else if op.Name == "modify" { // the file has beeen renamed or moved
		err := syncFile.RealFile.UpdateMetadata()
		if err != nil {
			return err
		}

		if syncFile.RealFile.GetHash() != syncFile.CopyFile.GetHash() { // if the file has changed
			err := syncFile.WriteRealToCopy()
			if err != nil {
				return err
			}
			err = syncFile.CopyFile.UpdateMetadata()
			if err != nil {
				return err
			}
		}
		if op.NewFile != "" {
			err := syncFile.CopyFile.Move(filepath.Join(PATH_COPY_FILES, Dir(op.NewFile)), filepath.Base(op.NewFile))
			if err != nil {
				return err
			}
		}
		syncFile.CopyFile.SetVersion(op.Version)
	}
	syncFile.lastOperation = op
	return nil
}

// when there is a chnage on another node, we apply it on the Copy file, then on the real file if possible
func (syncFile *SyncFileImpl) applyOperationFromOther(fileState FileState, conn ShosetConn) error {
	op := fileState.LastOperation
	if op.Version < syncFile.CopyFile.GetVersion() {
		// we ignore the operation because it is older than the one we already have
		// we send him the info of our file (shorter version)
		syncFile.FileTransfer.SendMessage(syncFile.getShortInfoMsg(), conn)
		//fmt.Println(fileState.UUID, ": upate file :", "nothing to do")
		return nil
	}
	if op.Version == syncFile.CopyFile.GetVersion() {
		if opMapScore[op.Name] < opMapScore[syncFile.lastOperation.Name] {
			// we ignore the operation because it is older than the one we already have
			// we send him the info of our file (short version)
			syncFile.FileTransfer.SendMessage(syncFile.getShortInfoMsg(), conn)
			//fmt.Println(fileState.UUID, ": upate file :", "nothing to do")
			return nil
		} else if opMapScore[op.Name] == opMapScore[syncFile.lastOperation.Name] {
			// this is the same type of operation, we keep the one with a smaller hash
			if op.Hash+op.NewFile > syncFile.lastOperation.Hash+syncFile.lastOperation.NewFile {
				// we ignore the operation because our hash of the last operation is smaller
				// we send him the info of our file (short version)
				syncFile.FileTransfer.SendMessage(syncFile.getShortInfoMsg(), conn)
				//fmt.Println(fileState.UUID, ": upate file :", "nothing to do")
				return nil
			} else if op.Hash+op.NewFile == syncFile.lastOperation.Hash+syncFile.lastOperation.NewFile {
				// do nothing because same version, same operation, same hash
				//fmt.Println(fileState.UUID, ": upate file :", "nothing to do")
				return nil
			}
		}
	}
	// if we arrive here, it means that we must apply the incoming operation
	if op.Name == "remove" {
		//syncFile.FileTransfer.Library.DeleteFile(syncFile.CopyFile.GetName())
		syncFile.FileTransfer.DeleteLeecher(syncFile.GetUUID())
		log.Println(fileState.UUID, ": upate file :", "remove file")
	} else if op.Name == "modify" { // the file has beeen renamed or moved

		if op.NewFile != filepath.Join(syncFile.CopyFile.GetRelativePath(), syncFile.CopyFile.GetName()) && op.NewFile != "" {
			// we rename directly the file
			err := syncFile.CopyFile.Move(filepath.Join(PATH_COPY_FILES, Dir(op.NewFile)), filepath.Base(op.NewFile))
			if err != nil {
				return err
			}
			err = syncFile.RealFile.Move(Dir(op.NewFile), filepath.Base(op.NewFile))
			if err != nil {
				// we move back the Copy file
				syncFile.CopyFile.Move(filepath.Join(PATH_COPY_FILES, Dir(op.File)), filepath.Base(op.File))
				return err
			}
		}
		if op.Hash != syncFile.CopyFile.GetHash() {
			if len(fileState.HashMap) != 0 {
				syncFile.CopyFile.SetHash(fileState.Hash)
				syncFile.CopyFile.SetHashMap(fileState.HashMap)
				syncFile.m.Unlock()
				syncFile.FileTransfer.InitLeecher(syncFile, conn)
				syncFile.m.Lock()
				log.Println(fileState.UUID, ": upate file :", "init leecher")
			} else {
				log.Println(fileState.UUID, ": upate file :", "asking info")
				go syncFile.FileTransfer.AskInfoFile(conn, fileState)
			}
		}
		syncFile.CopyFile.SetVersion(fileState.Version)
	}
	return nil
}

func (syncFile *SyncFileImpl) UpdateFile(fileState FileState, conn ShosetConn) error {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	//fmt.Println("comparing \n", conn.GetRemoteAddress(), fileState, "\n", conn.GetLocalAddress(), syncFile.CopyFile.GetVersion(), syncFile.CopyFile.GetHash(), syncFile.CopyFile.GetName(), syncFile.RealFile.GetRelativePath())
	var err error
	if (fileState.Version != syncFile.CopyFile.GetVersion()) || (fileState.Hash != syncFile.CopyFile.GetHash()) || (fileState.Name != syncFile.CopyFile.GetName()) || (fileState.Path != syncFile.RealFile.GetRelativePath()) {
		err = syncFile.applyOperationFromOther(fileState, conn)
		return err
	} else {
		if syncFile.status == "downloading" {
			leecher := syncFile.FileTransfer.GetLeecher(syncFile.uuid)
			if leecher != nil {
				syncFile.m.Unlock()
				leecher.InitDownload(conn)
				syncFile.m.Lock()
			} else {
				return fmt.Errorf("leecher not found when we are downloading")
			}

		}
		return nil
	}
}

func (syncFile *SyncFileImpl) SaveFileInfo() error {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	infoMap := syncFile.CopyFile.GetFileInfoMap()
	infoMap["uuid"] = syncFile.uuid
	fileInfo, err := json.Marshal(infoMap)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(syncFile.baseFilesDir, PATH_COPY_FILES, ".info", syncFile.uuid+".info"), fileInfo, 0644)
	if err != nil {
		log.Println("Error while saving file info: ", err)
		return err
	}
	return nil
}

func LoadSyncFileInfo(baseDir string, path string) (*SyncFileImpl, error) {
	openedFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(openedFile)
	fileInfoMap := make(map[string]interface{})
	err = decoder.Decode(&fileInfoMap)
	openedFile.Close()
	if err != nil {
		return nil, err
	}
	syncFile := NewSyncFile(baseDir)
	syncFile.uuid = fileInfoMap["uuid"].(string)
	CopyFile, err := LoadFileInfo(fileInfoMap)
	if err != nil {
		return nil, err
	}
	syncFile.CopyFile = CopyFile

	realFile, err := LoadFile(baseDir, CopyToRealPath(CopyFile.GetRelativePath()), CopyFile.GetName())
	if err != nil {
		return nil, err
	}
	syncFile.RealFile = realFile

	newHash, err := CopyFile.CalculateHash()
	if err != nil {
		return nil, err
	}
	if newHash != CopyFile.hash {
		// the file in the Copy has probably been modified
		// it occurs when the node is closed while a file is being downloaded
		CopyFile.version = 0
		return nil, err
	}
	CopyFile.UpdateMetadata()
	return syncFile, nil
}

// Getters and Setters

func (syncFile *SyncFileImpl) GetRealFile() File {
	return syncFile.RealFile
}

func (syncFile *SyncFileImpl) GetCopyFile() File {
	return syncFile.CopyFile
}

func (syncFile *SyncFileImpl) GetUUID() string {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	return syncFile.uuid
}

func (syncFile *SyncFileImpl) SetUUID(uuid string) {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	syncFile.uuid = uuid
}

func (syncFile *SyncFileImpl) SetFileTransfer(ft FileTransfer) {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	syncFile.FileTransfer = ft
}

func (syncFile *SyncFileImpl) GetLastOperation() Operation {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	return syncFile.lastOperation
}

func (syncFile *SyncFileImpl) getShortInfoMsg() msg.FileMessage {
	info := msg.FileMessage{MessageName: "sendInfo",
		FileUUID:      syncFile.uuid,
		FileName:      syncFile.CopyFile.GetName(),
		FileHash:      syncFile.CopyFile.GetHash(),
		FileVersion:   syncFile.CopyFile.GetVersion(),
		FilePath:      syncFile.RealFile.GetRelativePath(),
		FileOperation: syncFile.lastOperation.OperationMessage(),
	}
	info.InitMessageBase()
	return info
}

func (syncFile *SyncFileImpl) GetShortInfoMsg() msg.FileMessage {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	return syncFile.getShortInfoMsg()
}

// like shortInfo but we add the hashmap
func (syncFile *SyncFileImpl) GetFullInfoMsg() msg.FileMessage {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	info := syncFile.getShortInfoMsg()
	info.FileHashMap = syncFile.CopyFile.GetHashMap()
	info.FileSize = syncFile.CopyFile.GetSize()
	return info
}

func (syncFile *SyncFileImpl) GetFileState() FileState {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	return FileState{
		UUID:          syncFile.uuid,
		Name:          syncFile.CopyFile.GetName(),
		Hash:          syncFile.CopyFile.GetHash(),
		Version:       syncFile.CopyFile.GetVersion(),
		Path:          syncFile.RealFile.GetRelativePath(),
		LastOperation: syncFile.lastOperation,
	}
}

func (syncFile *SyncFileImpl) SetStatus(status string) {
	syncFile.m.Lock()
	defer syncFile.m.Unlock()
	syncFile.status = status
}

func (syncFile *SyncFileImpl) GetStatus() string {
	syncFile.m.RLock()
	defer syncFile.m.RUnlock()
	return syncFile.status
}
