package file_test

import (
	"os"
	"path/filepath"
	"testing"

	fileMod "github.com/ditrit/shoset/file"
)

func TestNewEmptyFile(t *testing.T) {
	workingDir := cleanDir(t)
	file, err := fileMod.NewEmptyFile(workingDir, "relativePath", "fileName", 10, "hash", 1, map[int]string{0: "hash"})
	if err != nil {
		t.Errorf("error in constructor : " + err.Error())
	}
	if file.GetRelativePath() != "relativePath" {
		t.Errorf("file relative path is not relativePath")
	}
	if file.GetName() != "fileName" {
		t.Errorf("file name is not fileName")
	}
	if file.GetSize() != 10 {
		t.Errorf("file size is not 10")
	}
	if file.GetHash() != "hash" {
		t.Errorf("file hash is not hash")
	}
	if file.GetVersion() != 1 {
		t.Errorf("file version is not 1")
	}
	if file.GetHashMap()[0] != "hash" {
		t.Errorf("file hashes is not hash")
	}

	fileInfo, err := os.Stat(filepath.Join(workingDir, "relativePath", "fileName"))
	if err != nil {
		t.Errorf("file not created")
	}
	if fileInfo.Size() != 10 {
		t.Errorf("file size is not 10")
	}
}

func TestLoadFile(t *testing.T) {
	workingDir := cleanDir(t)
	_, err := fileMod.LoadFile(workingDir, "relativePath", "fileName")
	// it normally create a file
	if err != nil {
		t.Errorf("error in constructor : " + err.Error())
	}
	_, err = os.Stat(filepath.Join(workingDir, "relativePath", "fileName"))
	if err != nil {
		t.Errorf("file not created : ")
	}
}

func TestLoadFileInfo(t *testing.T) {
	workingDir := cleanDir(t)
	err := os.MkdirAll(filepath.Join(workingDir, "relativePath"), 0755)
	if err != nil {
		t.Errorf("error in creation : " + err.Error())
	}
	fc, err := os.Create(filepath.Join(workingDir, "relativePath", "fileName"))
	if err != nil {
		t.Errorf("error in creation : " + err.Error())
	}
	fc.Close()

	fileInfo := make(map[string]interface{})
	fileInfo["baseFilesDir"] = workingDir
	fileInfo["relativePath"] = "relativePath"
	fileInfo["name"] = "fileName"
	fileInfo["size"] = float64(0)
	fileInfo["pieceSize"] = float64(256 * 1024)
	fileInfo["nbPieces"] = float64(0)
	fileInfo["hash"] = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	fileInfo["version"] = float64(1)
	fileInfo["hashMap"] = map[string]interface{}{}
	_, err = fileMod.LoadFileInfo(fileInfo)
	// it normally create a file
	if err != nil {
		t.Errorf("error in constructor : " + err.Error())
	}
	_, err = os.Stat(filepath.Join(workingDir, "relativePath", "fileName"))
	if err != nil {
		t.Errorf("file not created")
	}
}
