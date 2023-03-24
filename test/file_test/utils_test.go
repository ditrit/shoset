package file_test

import (
	"testing"

	fileMod "github.com/ditrit/shoset/file"
)

func TestRealToCopyPath(t *testing.T) {
	path := "hello/Downloads/1.txt"
	expected := ".copy/hello/Downloads/1.txt"
	result := fileMod.RealToCopyPath(path)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCopyToRealPath(t *testing.T) {
	path := ".copy/hello/Downloads/1.txt"
	expected := "hello/Downloads/1.txt"
	result := fileMod.CopyToRealPath(path)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
