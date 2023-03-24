package file_test

import (
	"fmt"
	"testing"

	mock "github.com/ditrit/shoset/test/mocks/file"
	"github.com/golang/mock/gomock"

	fileMod "github.com/ditrit/shoset/file"
)

func TestNewCommands(t *testing.T) {
	workingDir := cleanDir(t)
	createCopyDir(t)
	fmt.Println("workingDir", workingDir)

	ctrl := gomock.NewController(t)
	fileTransfer := mock.NewMockFileTransfer(ctrl)
	ec := fileMod.NewExternalCommands(fileTransfer)

	if ec.FileTransfer != fileTransfer {
		t.Errorf("NewExternalCommands() = %v, want %v", ec.FileTransfer, fileTransfer)
	}
}
