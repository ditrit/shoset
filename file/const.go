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
