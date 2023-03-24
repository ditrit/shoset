package fileSync

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog"
)

type ShosetConn interface {
	SendMessage(msg.Message) error
	GetLocalAddress() string
	GetRemoteAddress() string
	GetTCPInfo() (*syscall.TCPInfo, error)
	GetLogger() *zerolog.Logger
}

// contains range through a slice to search for a particular string
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func Max64(x int64, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func Min64(x int64, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

func Max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x int, y int) int {
	if x > y {
		return y
	}
	return x
}

func CalculateNumberOfChunks(fileSize int64, chunkSize int) int {
	return int(math.Ceil((float64(fileSize) / float64(chunkSize))))
}

// Give the optimal size of a piece of a file to have a better transfer rate
// The given size is between 256Kbytes and 1Mbytes (and given in bytes)
// The greater the file size, the greater the piece size
func CalculatePieceSize(fileSize int64) int {
	opt := int(fileSize / 1000)
	if opt <= 256*1024 {
		return 256 * 1024
	}
	if opt >= 1024*1024 {
		return 1024 * 1024
	}
	return 512 * 1024
}

func Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// convert a relative path from real to Copy (add .Copy before)
func RealToCopyPath(path string) string {
	listFolder := strings.Split(path, string(os.PathSeparator))
	listFolder = append([]string{PATH_COPY_FILES}, listFolder...)
	return filepath.Join(listFolder...)
}

// convert a relative path from real to Copy (add .Copy before)
func CopyToRealPath(path string) string {
	listFolder := strings.Split(path, string(os.PathSeparator))
	return filepath.Join(listFolder[1:]...)
}

// get the relative path from a full path, assuming baseDir is in the path
func PathToRelativePath(path string, baseDir string) string {
	listFolder := strings.Split(path, string(os.PathSeparator))
	listBaseDir := strings.Split(baseDir, string(os.PathSeparator))
	newList := listFolder[len(listBaseDir):]
	return filepath.Join(newList...)
}

// tells if the file/folder is in a subdirectory by comparing paths
func IsInDir(path string, dir string) bool {

	listFolder := strings.Split(path, string(os.PathSeparator))
	listDir := strings.Split(dir, string(os.PathSeparator))

	if len(listFolder) < len(listDir) {
		return false
	}
	for i := 0; i < len(listDir); i++ {
		if listFolder[i] != listDir[i] {
			return false
		}
	}
	return true
}

// get directory from a path
// like filepath.Dir but return "" instead of "." when no directory
func Dir(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	return dir
}
