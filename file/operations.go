package fileSync

import (
	"github.com/ditrit/shoset/msg"
)

/*
This file is used to describe an operation on a file
*/

type Operation struct {
	File    string // relative path and name of the file
	Name    string // name of the operation
	NewFile string // in case of a renamed or a moved file
	Version int    // version of the file
	Hash    string // hash of the file
}

// get the message format of an operation
func (operation Operation) OperationMessage() msg.OperationMessage {
	return msg.OperationMessage{
		Name:    operation.Name,
		File:    operation.File,
		NewFile: operation.NewFile,
		Version: operation.Version,
		Hash:    operation.Hash,
	}
}

// convert an operationMessage to an operation
func ToOperation(operation msg.OperationMessage) Operation {
	return Operation{
		File:    operation.File,
		Name:    operation.Name,
		NewFile: operation.NewFile,
		Version: operation.Version,
		Hash:    operation.Hash,
	}
}

func (operation Operation) String() string {
	if operation.NewFile != "" {
		return operation.Name + " " + operation.File + " into " + operation.NewFile
	} else {
		return operation.Name + " " + operation.File
	}
}
