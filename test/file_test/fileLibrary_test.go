package file_test

import (
	"os"
	"path/filepath"
	"testing"

	fileMod "github.com/ditrit/shoset/file"
)

// we assume that file and syncFile tests are ok

func TestNewFileLibrary(t *testing.T) {
	workingDir := cleanDir(t)
	fileMod.NewFileLibrary(workingDir)
}

func TestLoadLibrary(t *testing.T) {
	workingDir := cleanDir(t)
	createCopyDir(t)

	// we create a syncFile and store the info
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

	fileLibrary, err := fileMod.NewFileLibrary(workingDir) // it will automatically load the library
	if err != nil {
		t.Error(err)
	}
	_, err = fileLibrary.GetFile(uuid)
	if err != nil {
		t.Error(err)
	}

}

func TestUploadFile(t *testing.T) {
	workingDir := cleanDir(t)
	createCopyDir(t)

	// create the library
	fileLibrary, err := fileMod.NewFileLibrary(workingDir) // it will automatically load the library
	if err != nil {
		t.Error(err)
	}

	// we create a file
	_, err = os.Create(filepath.Join(workingDir, "fileName"))
	if err != nil {
		t.Error(err)
	}
	file, err := fileMod.LoadFile(workingDir, "", "fileName")
	if err != nil {
		t.Error(err)
	}

	syncFile, err := fileLibrary.UploadFile(file)
	if err != nil {
		t.Error(err)
	}
	if syncFile.GetRealFile() != file {
		t.Error("the real file is not the same")
	}

	// we check that the file is in the library
	_, err = fileLibrary.GetFile(syncFile.GetUUID())
	if err != nil {
		t.Error(err)
	}

	// the Copy file should have been created
	_, err = os.Stat(filepath.Join(workingDir, fileMod.PATH_COPY_FILES, "fileName"))
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateLibrary(t *testing.T) {}
