package fileSync

import (
	"fmt"
	"sync"
	"time"

	"github.com/ditrit/shoset/msg"
)

/*
This file is used to describe a file seeder
It is used when we have the full file and we want to share it
This small structure load the data from the wanted file and send it to the requester
*/

type FileSeeder struct {
	SyncFile         SyncFile
	File             File
	FileTransfer     FileTransfer
	DataAskedPerConn sync.Map //map[ShosetConn][int64]int // list of conn and the data they asked : to not send them twice the same data
	InterestedConn   sync.Map //map[ShosetConn]bool       // list of interested conn
	m                sync.Mutex
	lastSent         int64 // timestamp of last sent block : used to delete the seeder if no block is sent for a long time
}

func (fileSeeder *FileSeeder) InitSeeder(syncFile SyncFile, fileTransfer FileTransfer) {
	fileSeeder.SyncFile = syncFile
	fileSeeder.File = syncFile.GetCopyFile()
	fileSeeder.FileTransfer = fileTransfer
	fileSeeder.InterestedConn = sync.Map{}
	fileSeeder.lastSent = time.Now().UnixMilli()
}

// someone request a block of a piece
func (fileSeeder *FileSeeder) sendBlock(conn ShosetConn, begin int64, length int) error {
	// check if we didn't already send the block to the conn
	fileSeeder.m.Lock()
	defer fileSeeder.m.Unlock()
	beginsMap := make(map[int64]int)
	begins, ok := fileSeeder.DataAskedPerConn.Load(conn)
	if ok {
		beginsMap = begins.(map[int64]int)
		length2, ok := beginsMap[begin]
		if ok { // he already asked with this begin
			if length2 == length { // he already asked with this begin and this length
				return fmt.Errorf("already sent")
			}
		}
	}
	beginsMap[begin] = length
	fileSeeder.DataAskedPerConn.Store(conn, beginsMap)
	fileSeeder.lastSent = time.Now().UnixMilli()

	// load the data
	data, err := fileSeeder.File.LoadData(begin, length)
	if err != nil {
		return err
	}

	// send the block

	responseMsg := msg.FileMessage{
		MessageName: "sendChunk",
		FileUUID:    fileSeeder.SyncFile.GetUUID(),
		FileName:    fileSeeder.File.GetName(),
		FileHash:    fileSeeder.File.GetHash(),
		Begin:       begin,
		Length:      length,
		ChunkData:   data,
	}
	responseMsg.InitMessageBase()
	fileSeeder.FileTransfer.SendMessage(responseMsg, conn)
	return err
}

func (fileSeeder *FileSeeder) GetLastSent() int64 {
	fileSeeder.m.Lock()
	defer fileSeeder.m.Unlock()
	return fileSeeder.lastSent
}
