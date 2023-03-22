package fileSync

import (
	"io/fs"
)

// file
const (
	CHUNKSIZE       int         = 16 * 1024 // 16Kbytes
	CHUNKTIMEOUT    int64       = 1         //s
	PERMISSION      fs.FileMode = 0776      // permission for file and folders
	PATH_COPY_FILES string      = ".copy"
)

// Timeout Parameters
const (
	TIMEBEFOREANOTHERDECREASE   int64 = 3000 //ms
	TIMEOUTREQUEST              int64 = 2000 //ms
	TIMENOTASKINGDURINGDECREASE int64 = 3000 //ms
	TIMEBEFOREOPERATION         int64 = 1000 //ms
)
