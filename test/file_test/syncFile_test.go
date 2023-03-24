package file_test

// we assume that the tests from file_test.go are ok

import (
	"path/filepath"
	"testing"

	fileMod "github.com/ditrit/shoset/file"
)

func TestNewSyncFile(t *testing.T) {
	workingDir := cleanDir(t)
	fileMod.NewSyncFile(workingDir)
}

func TestWriting(t *testing.T) {
	workingDir := cleanDir(t)
	syncFile := fileMod.NewSyncFile(workingDir)
	CopyFile, err := fileMod.LoadFile(workingDir, filepath.Join(fileMod.PATH_COPY_FILES, "test"), "fileName")
	if err != nil {
		t.Error(err)
	}
	err = CopyFile.WriteChunk([]byte("hello"), 0)
	if err != nil {
		t.Error(err)
	}
	syncFile.CopyFile = CopyFile
	realFile, err := fileMod.LoadFile(workingDir, "test", "fileName")
	if err != nil {
		t.Error(err)
	}
	syncFile.RealFile = realFile

	// test writing Copy to real
	err = syncFile.WriteCopyToReal()
	if err != nil {
		t.Error(err)
	}
	hello, err := realFile.LoadData(0, 5)
	if err != nil {
		t.Error(err)
	}
	if string(hello) != "hello" {
		t.Error("file content is not correct : ", string(hello), " instead of hello")
	}

	// test writing real to Copy
	realFile.WriteChunk([]byte(" world"), 5)
	err = syncFile.WriteRealToCopy()
	if err != nil {
		t.Error(err)
	}

	helloWorld, err := CopyFile.LoadData(0, 11)
	if err != nil {
		t.Error(err)
	}
	if string(helloWorld) != "hello world" {
		t.Error("file content is not correct : ", string(helloWorld), " instead of hello world")
	}
}

func TestSaveFileInfo(t *testing.T) {
	workingDir := cleanDir(t)
	createCopyDir(t)
	syncFile := fileMod.NewSyncFile(workingDir)
	CopyFile, err := fileMod.LoadFile(workingDir, filepath.Join(fileMod.PATH_COPY_FILES, "test"), "fileName")
	if err != nil {
		t.Error(err)
	}
	err = CopyFile.WriteChunk([]byte("hello"), 0)
	if err != nil {
		t.Error(err)
	}
	syncFile.CopyFile = CopyFile
	realFile, err := fileMod.LoadFile(workingDir, "test", "fileName")
	if err != nil {
		t.Error(err)
	}
	syncFile.RealFile = realFile

	err = syncFile.SaveFileInfo()
	if err != nil {
		t.Error(err)
	}
}

func TestLoadSyncFileInfo(t *testing.T) {
	workingDir := cleanDir(t)
	createCopyDir(t)
	syncFile := fileMod.NewSyncFile(workingDir)
	CopyFile, err := fileMod.LoadFile(workingDir, filepath.Join(fileMod.PATH_COPY_FILES, "test"), "fileName")
	if err != nil {
		t.Error(err)
	}
	err = CopyFile.WriteChunk([]byte("hello"), 0)
	if err != nil {
		t.Error(err)
	}
	syncFile.CopyFile = CopyFile
	realFile, err := fileMod.LoadFile(workingDir, "test", "fileName")
	if err != nil {
		t.Error(err)
	}
	syncFile.RealFile = realFile

	// save the info first
	err = syncFile.SaveFileInfo()
	if err != nil {
		t.Error(err)
	}
	uuid := syncFile.GetUUID()

	syncFile2, err := fileMod.LoadSyncFileInfo(workingDir, filepath.Join(workingDir, fileMod.PATH_COPY_FILES, ".info", uuid+".info"))
	if err != nil {
		t.Error(err)
	}
	if syncFile2.CopyFile.GetName() != "fileName" {
		t.Error("file name is not correct")
	}
	if syncFile2.CopyFile.GetRelativePath() != filepath.Join(fileMod.PATH_COPY_FILES, "test") {
		t.Error("file relative path is not correct")
	}
	if syncFile2.RealFile.GetName() != "fileName" {
		t.Error("file name is not correct")
	}
	if syncFile2.RealFile.GetRelativePath() != "test" {
		t.Error("file relative path is not correct")
	}
}
