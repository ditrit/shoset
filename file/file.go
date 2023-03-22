package fileSync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

/*
This is used to do operations on one file (read, write data, calculate hash of the file, have the version, ...).
Each file in the library and in the copy of the library has a file instance.
*/

type File interface {
	LoadFromMap(fileInfoMap map[string]interface{})
	GetFileInfoMap() map[string]interface{}
	OpenFile() (*os.File, error)
	CloseFile()
	// write a chunk at the specified location
	WriteChunk(chunk []byte, offset int64) error
	String() string
	// Return the chunk of data at the specified index
	LoadData(chunk int64, chunkSize int) ([]byte, error)
	CalculateHashMap() (map[int]string, error)
	CalculateHash() (string, error)
	// update all the information about the file from what is on the disk
	UpdateMetadata() error
	// move the file to a new location (and / or rename it)
	Move(newRelativePath string, newName string) error
	GetName() string
	GetSize() int64
	GetHash() string
	GetHashMap() map[int]string
	GetHashChunk(index int) string
	GetPieceSize() int
	GetVersion() int
	SetVersion(version int)
	GetRelativePath() string
	SetRelativePath(relativePath string)
	SetName(name string)
	SetHash(hash string)
	SetHashMap(hashMap map[int]string)
}

type FileImpl struct {
	name         string
	baseFilesDir string // path to the directory where all the files are stored
	relativePath string // path to the file relative to the baseFilesDir
	size         int64  // size in bytes of the file
	pieceSize    int    // size in bytes of a piece of the file
	nbPieces     int    // number of pieces of the file
	hash         string
	hashMap      map[int]string
	version      int // version of the file

	openedFile       *os.File
	openedTime       int64
	closingGoRoutine bool

	m sync.RWMutex
}

// We have a new file coming from an other node : it creates a new file
func NewEmptyFile(baseFilesDir string, relativePath string, name string, size int64, hash string, version int, hashMap map[int]string) (*FileImpl, error) {
	file := FileImpl{}
	file.m.Lock()
	defer file.m.Unlock()

	file.baseFilesDir = baseFilesDir
	file.relativePath = relativePath
	file.name = name
	file.size = size
	file.pieceSize = CalculatePieceSize(size)
	file.hashMap = make(map[int]string, file.pieceSize)
	file.nbPieces = int(math.Ceil(float64(file.size) / float64(file.pieceSize)))
	if len(hashMap) != file.nbPieces { // there is a problem in the way the hashMap is given
		return &file, fmt.Errorf("the hashMap has a size of " + strconv.Itoa(len(hashMap)) + " whereas the file has " + strconv.Itoa(file.nbPieces) + " pieces")
	}
	file.version = version
	file.hash = hash

	for i := 0; i < file.nbPieces; i++ {
		file.hashMap[i] = hashMap[i]
	}

	// create the directory in the relativePath if it doesn't exist
	err := os.MkdirAll(filepath.Join(baseFilesDir, file.relativePath), PERMISSION)
	if err != nil && !os.IsExist(err) {
		return &file, err
	}
	// create the file in the relative directory
	fc, err := os.Create(filepath.Join(baseFilesDir, file.relativePath, file.name))
	if err != nil { // if there is a problem with the file creation
		return &file, err
	}
	fc.Close()
	// create the file and fill it with empty bytes (in the .copy directory)
	fd, err := os.Create(filepath.Join(baseFilesDir, file.relativePath, file.name))
	if err != nil { // if there is a problem with the file creation
		return &file, err
	}
	_, err = fd.Seek(int64(size-1), 0)
	if err != nil {
		return &file, err
	}
	_, err = fd.Write([]byte{0})
	if err != nil {
		return &file, err
	}
	err = fd.Close()
	if err != nil {
		return &file, err
	}
	return &file, err
}

// We load a file (if it doesn't exist, we create it) and load the metadata from it (size, hash, hashMap, ...)
func LoadFile(baseFilesDir string, relativePath string, name string) (*FileImpl, error) {
	file := FileImpl{}
	file.m.Lock()
	file.baseFilesDir = baseFilesDir
	file.relativePath = relativePath
	file.name = name
	file.version = 0
	file.m.Unlock()
	// we create the file if it doesn't exist
	_, err := os.Stat(filepath.Join(baseFilesDir, relativePath, name))
	if err != nil {
		if os.IsNotExist(err) {
			// create the directory in the relativePath if it doesn't exist
			err := os.MkdirAll(filepath.Join(baseFilesDir, relativePath), PERMISSION)
			if err != nil && !os.IsExist(err) {
				return &file, err
			}
			// create the file in the relative directory
			fc, err := os.Create(filepath.Join(baseFilesDir, relativePath, name))
			if err != nil { // if there is a problem with the file creation
				return &file, err
			}
			fc.Close()
		}
	}
	err = file.UpdateMetadata()
	return &file, err
}

// Load the file information from the info on the map and detect differences with the actual file
// it is used when we launch the node to detect if some files have been modified accidentally
func LoadFileInfo(fileInfoMap map[string]interface{}) (*FileImpl, error) {
	file := FileImpl{}

	file.LoadFromMap(fileInfoMap)

	file.m.Lock()
	defer file.m.Unlock()
	newHash, err := file.calculateHash()
	if err != nil {
		return nil, err
	}
	if newHash != file.hash {
		// the file in the .copy has probably been modified
		// it occurs when the node is closed while a file is being downloaded
		file.version = 0
		return nil, fmt.Errorf("the file has been modified, hashes are different")
	}
	return &file, nil
}

func (file *FileImpl) LoadFromMap(fileInfoMap map[string]interface{}) {
	file.m.Lock()
	defer file.m.Unlock()

	file.baseFilesDir = fileInfoMap["baseFilesDir"].(string)
	file.relativePath = fileInfoMap["relativePath"].(string)
	file.name = fileInfoMap["name"].(string)
	file.size = int64(fileInfoMap["size"].(float64))
	file.pieceSize = int(fileInfoMap["pieceSize"].(float64))
	file.nbPieces = int(fileInfoMap["nbPieces"].(float64))
	file.hash = fileInfoMap["hash"].(string)
	newMap := fileInfoMap["hashMap"].(map[string]interface{})
	for i, hash := range newMap {
		stri, _ := strconv.Atoi(i)
		file.hashMap[stri] = hash.(string)
	}
	file.version = int(fileInfoMap["version"].(float64))
}

func (file *FileImpl) GetFileInfoMap() map[string]interface{} {
	file.m.RLock()
	defer file.m.RUnlock()
	fileInfo := make(map[string]interface{})
	fileInfo["name"] = file.name
	fileInfo["baseFilesDir"] = file.baseFilesDir
	fileInfo["relativePath"] = file.relativePath
	fileInfo["size"] = file.size
	fileInfo["pieceSize"] = file.pieceSize
	fileInfo["nbPieces"] = file.nbPieces
	fileInfo["hash"] = file.hash
	newHashMap := make(map[int]string)
	for i := 0; i < file.nbPieces; i++ {
		newHashMap[i] = file.hashMap[i]
	}
	fileInfo["hashMap"] = newHashMap
	fileInfo["version"] = file.version
	return fileInfo
}

func (file *FileImpl) openFile() (*os.File, error) {
	if file.openedFile == nil { // if we don't have already opened the file
		path := filepath.Join(file.baseFilesDir, file.relativePath, file.name)
		openedFile, err := os.OpenFile(path, os.O_RDWR, PERMISSION)
		if err != nil {
			return nil, err
		}
		file.openedFile = openedFile
		file.openedTime = time.Now().Unix()
		if !file.closingGoRoutine {
			go file.closeFileAfterTimeout()
		}
	}
	return file.openedFile, nil
}

// to execute in a gori=outine : close the file after 5s if nobody have written or read from the file
func (file *FileImpl) closeFileAfterTimeout() {
	for {
		file.m.Lock()
		file.closingGoRoutine = true
		if file.openedFile != nil && (time.Now().Unix()-file.openedTime) > 5 { // if the file has been opened for more than 5s and nobody have made a request for a chunk
			file.closeFile()
			file.closingGoRoutine = false
			file.m.Unlock()
			break
		}
		file.m.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (file *FileImpl) OpenFile() (*os.File, error) {
	file.m.Lock()
	defer file.m.Unlock()
	return file.openFile()
}

func (file *FileImpl) closeFile() {
	if file.openedFile != nil {
		file.openedFile.Close()
		file.openedFile = nil
	}
}

func (file *FileImpl) CloseFile() {
	file.m.Lock()
	defer file.m.Unlock()
	file.closeFile()
}

// write a chunk at the specified location
func (file *FileImpl) WriteChunk(chunk []byte, offset int64) error {
	file.m.Lock()
	defer file.m.Unlock()
	openedFile, err := file.openFile()
	if err != nil {
		return err
	}

	n, err := openedFile.WriteAt(chunk, offset)
	if err != nil {
		return err
	}
	if n != len(chunk) {
		return fmt.Errorf("error while writing chunk")
	}
	return nil
}

func (file *FileImpl) String() string {
	file.m.RLock()
	defer file.m.RUnlock()
	var result string
	result += "Name (file) : " + file.name + "\n"
	result += "Path of the library : " + file.baseFilesDir + "\n"
	result += "RelativePath : " + file.relativePath + "\n"
	result += "Size: " + fmt.Sprint(file.size) + "\n"

	return result
}

// Return the chunk of data at the specified index
func (file *FileImpl) LoadData(chunk int64, chunkSize int) ([]byte, error) {
	file.m.Lock()
	defer file.m.Unlock()
	//fmt.Println("LoadData", chunk, chunkSize)
	return file.loadData(chunk, chunkSize)
}

// Return the chunk of data at the specified index
func (file *FileImpl) loadData(begin int64, size int) ([]byte, error) {
	dataToRead := Min64(int64(size), file.size-begin)
	if dataToRead <= 0 {
		return []byte{}, nil
	}
	data := make([]byte, dataToRead)
	openedFile, err := file.openFile()
	if err != nil {
		return data, err
	}

	_, err = openedFile.ReadAt(data, int64(begin)) // read the file from the offset chunk*chunkSize
	if err != nil && err != io.EOF {
		return data, err
	}

	return data, nil
}

func (file *FileImpl) calculateHashMap() (map[int]string, error) {
	// we do calculate hash chunk by chunk, add them together and then do a hash of the list of hash
	// this way we can check the hash of a file without having to load it in memory
	size, err := file.readSize()
	if err != nil {

		return nil, err
	}
	nbPieces := int(size / int64(file.pieceSize))
	hashMap := make(map[int]string, nbPieces)
	for i := 0; i < nbPieces; i++ {
		chunk, err := file.loadData(int64(i)*int64(file.pieceSize), file.pieceSize)
		if err != nil {
			return nil, err
		}
		hashMap[i] = Hash(chunk)
	}
	return hashMap, nil
}

func (file *FileImpl) CalculateHashMap() (map[int]string, error) {
	file.m.Lock()
	defer file.m.Unlock()
	return file.calculateHashMap()
}

func (file *FileImpl) calculateHash() (string, error) {
	hashMap, err := file.calculateHashMap()
	if err != nil {
		return "", err
	}

	hasher := sha256.New()
	for i := 0; i < file.nbPieces; i++ {
		hasher.Write([]byte(hashMap[i]))
	}
	//hasher.Write([]byte(file.name))
	//hasher.Write([]byte([]byte(strconv.Itoa(file.version))))
	hex.EncodeToString(hasher.Sum(nil))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (file *FileImpl) CalculateHash() (string, error) {
	file.m.Lock()
	defer file.m.Unlock()
	return file.calculateHash()
}

// read the size of the file
func (file *FileImpl) readSize() (int64, error) {
	info, err := os.Stat(filepath.Join(file.baseFilesDir, file.relativePath, file.name))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// update all the information about the file from what is on the disk
func (file *FileImpl) UpdateMetadata() error {
	file.m.Lock()
	defer file.m.Unlock()
	size, err := file.readSize()
	if err != nil {
		return err
	}
	file.size = size
	file.pieceSize = CalculatePieceSize(size)
	file.nbPieces = int(size / int64(file.pieceSize))
	file.hashMap = make(map[int]string, file.pieceSize)
	file.nbPieces = int(math.Ceil(float64(file.size) / float64(file.pieceSize)))

	file.hashMap, err = file.calculateHashMap()
	if err != nil {
		return err
	}
	file.hash, err = file.calculateHash()
	if err != nil {
		return err
	}
	return nil
}

// move the file to a new location (and / or rename it)
func (file *FileImpl) Move(newRelativePath string, newName string) error {
	file.m.Lock()
	defer file.m.Unlock()
	if file.openedFile != nil {
		file.openedFile.Close()
	}
	err := os.Rename(filepath.Join(file.baseFilesDir, file.relativePath, file.name), filepath.Join(file.baseFilesDir, newRelativePath, newName))
	if err != nil {
		return err
	}
	file.relativePath = newRelativePath
	file.name = newName
	return nil
}

/*
setters and getters
*/

func (file *FileImpl) GetName() string {
	file.m.RLock()
	defer file.m.RUnlock()

	return file.name
}

func (file *FileImpl) GetSize() int64 {
	file.m.RLock()
	defer file.m.RUnlock()

	return file.size
}

func (file *FileImpl) GetHash() string {
	file.m.RLock()
	defer file.m.RUnlock()

	return file.hash
}

func (file *FileImpl) GetHashMap() map[int]string {
	file.m.RLock()
	defer file.m.RUnlock()
	newHashMap := make(map[int]string)
	for k, v := range file.hashMap {
		newHashMap[k] = v
	}
	return newHashMap
}

func (file *FileImpl) GetHashChunk(index int) string {
	file.m.RLock()
	defer file.m.RUnlock()
	return file.hashMap[index]
}

func (file *FileImpl) GetPieceSize() int {
	file.m.RLock()
	defer file.m.RUnlock()
	return file.pieceSize
}

func (file *FileImpl) GetVersion() int {
	file.m.RLock()
	defer file.m.RUnlock()
	return file.version
}

func (file *FileImpl) SetVersion(version int) {
	file.m.Lock()
	defer file.m.Unlock()
	file.version = version
}

func (file *FileImpl) GetRelativePath() string {
	file.m.RLock()
	defer file.m.RUnlock()
	return file.relativePath
}

func (file *FileImpl) SetRelativePath(relativePath string) {
	file.m.Lock()
	defer file.m.Unlock()
	err := os.MkdirAll(filepath.Join(file.baseFilesDir, relativePath), 0755)
	if err != nil {
		log.Println("Error creating directory", err)
	}
	file.relativePath = relativePath
}

func (file *FileImpl) SetName(name string) {
	file.m.Lock()
	defer file.m.Unlock()
	file.name = name
}

func (file *FileImpl) SetHash(hash string) {
	file.m.Lock()
	defer file.m.Unlock()
	file.hash = hash
}

func (file *FileImpl) SetHashMap(hashMap map[int]string) {
	file.m.Lock()
	defer file.m.Unlock()
	for k, v := range hashMap {
		file.hashMap[k] = v
	}
}
