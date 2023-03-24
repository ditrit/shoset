package file_test

import (
	"os"
	"path/filepath"
	"testing"
)

func getBaseDir(t *testing.T) string {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Errorf(err.Error())
	}
	if filepath.Base(workingDir) == "file_test" {
		workingDir = filepath.Dir(workingDir)
	}
	baseDir := filepath.Join(workingDir, "testRep")
	return baseDir
}

func cleanDir(t *testing.T) string {
	baseDir := getBaseDir(t)

	os.RemoveAll(baseDir)
	os.MkdirAll(baseDir, 0755)
	return baseDir
}

func createCopyDir(t *testing.T) {
	workingDir := getBaseDir(t)
	CopyDir := filepath.Join(workingDir, ".copy", ".info")
	os.MkdirAll(CopyDir, 0755)
}
