package fileSync

import (
	"sync"
	"time"

	"github.com/ditrit/shoset/msg"
)

type Piece struct {
	SyncFile          SyncFile
	File              File
	pieceId           int
	hash              string            // hash of the piece
	pieceSize         int               // size in byte of the piece
	blockSize         int               // size in byte of a block
	blocks            [][CHUNKSIZE]byte // blocks of the piece
	blockPossessed    []bool            // bitfield of the blocks we have
	nbBlocksPossessed int               // number of blocks we have
	blockRequested    []bool            // list of the blocks we have already requested
	blockTimestamp    []int64           // timestamp of the last request of the block
	nbBlocks          int               // number of blocks in the piece
	noBlock           bool              // true if we download the piece at one time (we are not using blocks)
	data              []byte            // data of the piece
	pieceTimestamp    int64             // timestamp of the last request of the piece

	ConnInformation ConnInfo // information about the connection we are downloading the piece from

	m sync.Mutex
}

func NewPiece(syncFile SyncFile, pieceId int, pieceSize int, hash string, connInfo ConnInfo) *Piece {
	var piece Piece
	piece.SyncFile = syncFile
	piece.File = syncFile.GetCopyFile()
	piece.pieceId = pieceId
	piece.pieceSize = pieceSize
	piece.hash = hash
	piece.blockSize = CHUNKSIZE
	piece.nbBlocks = pieceSize / CHUNKSIZE
	piece.blocks = make([][CHUNKSIZE]byte, piece.nbBlocks)
	piece.blockPossessed = make([]bool, piece.nbBlocks)
	piece.nbBlocksPossessed = 0
	piece.blockRequested = make([]bool, piece.nbBlocks) // false by default
	piece.blockTimestamp = make([]int64, piece.nbBlocks)
	piece.ConnInformation = connInfo
	return &piece
}

func (piece *Piece) GetNoBlock() bool {
	piece.m.Lock()
	defer piece.m.Unlock()
	return piece.noBlock
}

func (piece *Piece) SetNoBlock(noBlock bool) {
	piece.m.Lock()
	defer piece.m.Unlock()
	piece.noBlock = noBlock
}

func (piece *Piece) SetData(data []byte) {
	piece.m.Lock()
	defer piece.m.Unlock()
	piece.data = data
}

func (piece *Piece) AddBlockToPiece(blockId int, block []byte) bool {
	piece.m.Lock()
	defer piece.m.Unlock()
	if piece.blockPossessed[blockId] {
		return false
	}
	copy(piece.blocks[blockId][:], block)
	piece.blockPossessed[blockId] = true
	piece.blockRequested[blockId] = false
	piece.nbBlocksPossessed++
	return true
}

func (piece *Piece) IsComplete() bool {
	piece.m.Lock()
	defer piece.m.Unlock()
	return piece.nbBlocksPossessed == piece.nbBlocks
}

func (piece *Piece) GetId() int {
	piece.m.Lock()
	defer piece.m.Unlock()
	return piece.pieceId
}

func (piece *Piece) getNextRequestedBlock() int {
	for i, blockPossessed := range piece.blockPossessed {
		if !blockPossessed && !piece.blockRequested[i] {
			piece.blockRequested[i] = true
			piece.blockTimestamp[i] = time.Now().UnixMilli()
			return i
		}
	}
	return -1
}

func (piece *Piece) WritePiece() error {
	piece.m.Lock()
	defer piece.m.Unlock()
	if len(piece.data) > 0 { // if the piece is entirely downloaded in onte itime and the data is in piece.data
		err := piece.File.WriteChunk(piece.data, int64(piece.pieceId)*int64(piece.pieceSize))
		return err
	}
	data := []byte{}
	for i := 0; i < piece.nbBlocks; i++ {
		data = append(data, piece.blocks[i][:]...)
	}
	return piece.File.WriteChunk(data, int64(piece.pieceId)*int64(piece.pieceSize))
}

// return true if we made all the block requests and the connInformation say that we can still ask for more, false otherwise
func (piece *Piece) downloadNextBlocks() bool {
	piece.m.Lock()
	count := 0
	if piece.noBlock {
		return false
	}
	for piece.ConnInformation.CanRequestBlock() && count < 2 {

		blockId := piece.getNextRequestedBlock()
		if blockId == -1 { // we reached the end of the piece, there is no more block to requests
			piece.m.Unlock()
			//piece.ConnInformation.Leecher.DownloadNextPieces(*piece.ConnInformation.Conn)
			return true
		}
		piece.SyncFile.GetUUID()
		piece.File.GetName()
		message := msg.FileMessage{
			MessageName: "askChunk",
			FileUUID:    piece.SyncFile.GetUUID(),
			FileName:    piece.File.GetName(),
			FileHash:    piece.File.GetHash(),
			Begin:       int64(piece.pieceId)*int64(piece.pieceSize) + int64(blockId*piece.blockSize),
			Length:      piece.blockSize,
		}
		message.InitMessageBase()
		//TODO : sned the message
		piece.ConnInformation.AddRequestBlock()
		//fmt.Println(conn.GetLocalAddress(), "asking piece", piece.pieceId, "to", conn.GetRemoteAddress())
		count++
	}
	piece.m.Unlock()
	return false
}

func (piece *Piece) downloadPiece() {
	piece.m.Lock()
	defer piece.m.Unlock()
	if piece.noBlock {
		//fmt.Println(conn.GetLocalAddress(), "asking piece", piece.pieceId, "to", conn.GetRemoteAddress())
		piece.pieceTimestamp = time.Now().UnixMilli()
		message := msg.FileMessage{
			MessageName: "askChunk",
			FileUUID:    piece.SyncFile.GetUUID(),
			FileName:    piece.File.GetName(),
			FileHash:    piece.File.GetHash(),
			Begin:       int64(piece.pieceId) * int64(piece.pieceSize),
			Length:      piece.pieceSize,
		}
		message.InitMessageBase()
		//TODO : send the message
	}
}

// return true if a timeout occured
func (piece *Piece) CheckRequestTimeout() bool {
	piece.m.Lock()
	timeoutOccured := false
	conn := piece.ConnInformation.GetConn()
	conn.GetLogger().Trace().Msgf("%v timeout check for piece %d", conn.GetLocalAddress(), piece.pieceId)
	if piece.noBlock {
		if time.Now().UnixMilli()-piece.pieceTimestamp > CHUNKTIMEOUT {
			conn := piece.ConnInformation.GetConn()
			conn.GetLogger().Trace().Msgf("%v --------------------- timeout for piece %d / %d ---------------- asked to %v", conn.GetLocalAddress(), piece.pieceId, int64(piece.pieceId)*int64(piece.pieceSize), conn.GetRemoteAddress())
			piece.m.Unlock()
			piece.ConnInformation.DecreaseLevel()
			piece.ConnInformation.GetLeecher().AddToRequestPriority(piece.pieceId, piece.ConnInformation)
			piece.m.Lock()
			timeoutOccured = true
		}
	} else {
		for i, blockPossessed := range piece.blockPossessed {
			if !blockPossessed && piece.blockRequested[i] {
				if time.Now().UnixMilli()-piece.blockTimestamp[i] > CHUNKTIMEOUT {
					conn := piece.ConnInformation.GetConn()
					conn.GetLogger().Trace().Msgf("%v --------------------- timeout for piece %d and block %d / %d ---------------- asked to %v", conn.GetLocalAddress(), piece.pieceId, i, int64(piece.pieceId)*int64(piece.pieceSize), conn.GetRemoteAddress())
					go func() {
						// we remove this request 3s after, if we didn't received it
						time.Sleep(3 * time.Second)
						piece.m.Lock()
						if !piece.blockPossessed[i] && piece.blockRequested[i] {
							conn.GetLogger().Trace().Msgf("reasking %d and block %d / %d ---------------- asked to %v", piece.pieceId, i, int64(piece.pieceId)*int64(piece.pieceSize)+int64(i*piece.blockSize), conn.GetRemoteAddress())
							piece.blockRequested[i] = false
							piece.m.Unlock()
							piece.ConnInformation.RemoveRequestBlock()
							piece.ConnInformation.DecreaseLevel()
						} else {
							piece.m.Unlock()
						}
					}()
					timeoutOccured = true
				}
				if time.Now().UnixMilli()-piece.blockTimestamp[i] > 10000 { // it's been 10s since we asked for this block, we remove the request for the entire piece
					piece.m.Unlock()
					piece.ConnInformation.DecreaseLevel()
					piece.ConnInformation.GetLeecher().AddToRequestPriority(piece.pieceId, piece.ConnInformation)
					piece.m.Lock()
					for i, blockPossessed := range piece.blockPossessed {
						if !blockPossessed && piece.blockRequested[i] {
							piece.blockRequested[i] = false
							piece.ConnInformation.RemoveRequestBlock()
						}
					}
					break
				}
			}
		}
	}
	piece.m.Unlock()
	return timeoutOccured
}

func (piece *Piece) CheckHash() bool {
	piece.m.Lock()
	defer piece.m.Unlock()
	if len(piece.data) > 0 {
		return piece.hash == Hash(piece.data)
	}

	data := []byte{}
	for i := 0; i < piece.nbBlocks; i++ {
		data = append(data, piece.blocks[i][:]...)
	}
	if piece.hash != Hash(data) {
		conn := piece.ConnInformation.GetConn()
		conn.GetLogger().Trace().Msgf("%d hash not matching %v : %v", piece.pieceId, piece.hash, Hash(data))
	}
	return piece.hash == Hash(data)
}

func (piece *Piece) SetConnInfo(connInfo ConnInfo) {
	piece.m.Lock()
	defer piece.m.Unlock()
	piece.ConnInformation = connInfo
}
