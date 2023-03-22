package fileSync

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ditrit/shoset/msg"
)

/*
This file is used to describe a file leecher
a file leecher is a structure that download a file from a file seeder and share it at the same time
It inherits from the file seeder
This is inspired by the bittorrent protocol
*/

type FileLeecher struct {
	FileSeeder
	hash                    string         // hash of the file we are downloading
	hashMap                 map[int]string // map of the hash of the pieces of the file
	nbPieces                int            // total number of pieces in the file
	nbPiecesReceived        int
	pieceSize               int                     // size (in bytes) of a piece
	nbBlocks                int                     // number of blocks per piece
	receivedPieces          []bool                  // pieces we have received
	requestedPieces         []bool                  // pieces we have requested
	piecesBeingRequested    map[int]*Piece          // pieces that are being requested
	pieceAskedToConn        map[int]ConnInfo        // map of the piece we have asked to a connection
	piecesToRequestPriority map[int]bool            // pieces that we want to request with top prioriyt because unfinished pieces. All these pieces are also in piecesBeingRequested
	pieceRarity             []int                   // rarity of the pieces (nb of peer that have the piece)
	ConnInfoMap             map[ShosetConn]ConnInfo // give us the info of a connection
	GetBitfiledFile         map[ShosetConn]bool     // map of conn we have asked for bitfield

	stop           chan bool // channel to stop the leecher
	stopped        bool      // true if the leecher is stopped
	startTimestamp int64     // timestamp of the start of the leecher
}

func NewFileLeecher(syncFile SyncFile, fileTransfer FileTransfer) *FileLeecher {
	var fileLeecher FileLeecher
	fileLeecher.SyncFile = syncFile
	fileLeecher.InitSeeder(syncFile, fileTransfer)
	fileLeecher.hash = syncFile.GetCopyFile().GetHash()
	fileLeecher.hashMap = syncFile.GetCopyFile().GetHashMap()
	fileLeecher.pieceSize = fileLeecher.File.GetPieceSize()
	fileLeecher.nbPieces = len(fileLeecher.hashMap)
	fileLeecher.nbPiecesReceived = 0
	fileLeecher.nbBlocks = int(fileLeecher.File.GetPieceSize()) / CHUNKSIZE

	fileLeecher.receivedPieces = make([]bool, fileLeecher.nbPieces)
	fileLeecher.requestedPieces = make([]bool, fileLeecher.nbPieces)
	fileLeecher.piecesBeingRequested = make(map[int]*Piece)
	fileLeecher.pieceAskedToConn = make(map[int]ConnInfo)
	fileLeecher.piecesToRequestPriority = make(map[int]bool)
	fileLeecher.pieceRarity = make([]int, fileLeecher.nbPieces)

	fileLeecher.stop = make(chan bool)
	fileLeecher.ConnInfoMap = make(map[ShosetConn]ConnInfo)
	fileLeecher.GetBitfiledFile = make(map[ShosetConn]bool)

	fileLeecher.startTimestamp = time.Now().UnixMilli()

	rand.Seed(time.Now().UnixNano())
	return &fileLeecher
}

// a newer version has appeared during the download
func (fileLeecher *FileLeecher) UpdateLeeching() {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()

	// update the receiveMap in a clever way
	// we don't want to loose the pieces we have already received
	newHash := fileLeecher.SyncFile.GetCopyFile().GetHash()
	if newHash != fileLeecher.hash {
		newHashMap := fileLeecher.SyncFile.GetCopyFile().GetHashMap()
		newReceivedPieces := make([]bool, len(newHashMap))

		for i, newHash := range newHashMap {
			hash, ok := fileLeecher.hashMap[i]
			if ok && hash == newHash { // if the same piece
				newReceivedPieces[i] = fileLeecher.receivedPieces[i]
			}
		}

		fileLeecher.hashMap = newHashMap
		fileLeecher.hash = newHash
		fileLeecher.receivedPieces = newReceivedPieces
		fileLeecher.nbPieces = len(newHashMap)
		fileLeecher.nbPiecesReceived = 0
		fileLeecher.requestedPieces = make([]bool, fileLeecher.nbPieces)
		fileLeecher.piecesBeingRequested = make(map[int]*Piece)
		fileLeecher.pieceAskedToConn = make(map[int]ConnInfo)
		fileLeecher.piecesToRequestPriority = make(map[int]bool)
		fileLeecher.pieceRarity = make([]int, fileLeecher.nbPieces)

		fileLeecher.ConnInfoMap = make(map[ShosetConn]ConnInfo)
		fileLeecher.GetBitfiledFile = make(map[ShosetConn]bool)
	}
}

func (fileLeecher *FileLeecher) SendHaveMessage(pieceId int) {
	haveMessage := msg.FileMessage{
		MessageName: "have",
		FileUUID:    fileLeecher.SyncFile.GetUUID(),
		FileName:    fileLeecher.File.GetName(),
		FileHash:    fileLeecher.File.GetHash(),
		PieceNumber: pieceId,
	}
	haveMessage.InitMessageBase()
	fileLeecher.InterestedConn.Range(func(key, value interface{}) bool {
		conn := key.(ShosetConn)
		if value.(bool) {
			fileLeecher.FileTransfer.SendMessage(haveMessage, conn)
		}
		return true
	})
}

func (fileLeecher *FileLeecher) ReceiveBitfieldMessage(conn ShosetConn, bitfield []bool) {
	fileLeecher.m.Lock()
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	fileLeecher.m.Unlock()
	if !ok { // new connection
		fileLeecher.InitDownload(conn)
		return
	}
	if connInfo.IsReady() { // if we already have the bitfield
		return
	}
	fileLeecher.m.Lock()
	delete(fileLeecher.GetBitfiledFile, conn)
	if bitfield == nil {
		bitfield = make([]bool, fileLeecher.nbPieces)
	}
	connInfo.SetBitfield(bitfield)
	for i, piece := range bitfield {
		if piece {
			fileLeecher.pieceRarity[i] += 1
		}
	}
	fileLeecher.m.Unlock()
	fileLeecher.DownloadNextPieces(conn)
}

func (fileLeecher *FileLeecher) ReceiveHaveMessage(conn ShosetConn, pieceId int) {
	fileLeecher.m.Lock()
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if !ok || !connInfo.IsReady() { // new connection
		fileLeecher.launchDownload(conn)
		fileLeecher.m.Unlock()
		return
	}
	if !connInfo.HavePiece(pieceId) && !fileLeecher.receivedPieces[pieceId] { // if we did'nt know he had this piece and we don't have it
		connInfo.AddPiece(pieceId)
		fileLeecher.pieceRarity[pieceId]++
		// we can ask for this piece
		fileLeecher.m.Unlock()
		fileLeecher.DownloadNextPieces(conn)
	} else {
		fileLeecher.m.Unlock()
	}
}

// we remove a conn from the authorised conn for a certain time
func (fileLeecher *FileLeecher) removeConn(conn ShosetConn) {
	conn.GetLogger().Info().Msg("------- RemoveConn ------- from " + conn.GetLocalAddress() + " to " + conn.GetRemoteAddress())

	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if !ok {
		return
	}
	available, _, _ := connInfo.GetAvailable()
	if available {
		connInfo.SetAvailable(false)

		for _, pieceId := range connInfo.GetPiecesRequested() {
			fileLeecher.piecesToRequestPriority[pieceId] = true
			connInfo.RemovePieceRequested(pieceId)
		}

		_, inc, _ := connInfo.GetAvailable()
		go func() {
			conn.GetLogger().Info().Msgf("sleeping %d milliseconds before downloading again", inc)
			time.Sleep(time.Duration(inc*(rand.Intn(10)+1)) * time.Millisecond)
			connInfo.SetAvailable(true)
			conn.GetLogger().Info().Msg("sleeping finished, downloading again")
			fileLeecher.DownloadNextPieces(conn)
		}()
	}
}

func (fileLeecher *FileLeecher) ReceiveChunk(conn ShosetConn, begin int64, length int, block []byte) error {
	fileLeecher.m.Lock()
	if fileLeecher.stopped {
		fileLeecher.m.Unlock()
		return fmt.Errorf("file leecher is stopped")
	}
	// retrieve the connection info
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if !ok { // new connection
		fileLeecher.launchDownload(conn)
		fileLeecher.m.Unlock()
		return fmt.Errorf("connection %s is not in the map", conn.GetRemoteAddress())
	}
	// retrieve the piece
	pieceId := int(begin / int64(fileLeecher.pieceSize))
	piece, ok := fileLeecher.piecesBeingRequested[pieceId]
	if !ok { // this piece is not currently being downloaded
		fileLeecher.m.Unlock()
		return fmt.Errorf("piece %d is not currently being downloaded", pieceId)
	}

	connInfo.AddBytesSeries(len(block))

	fileLeecher.FileTransfer.DecreaseMissingLength(len(block))

	if length == fileLeecher.pieceSize {
		// if the piece is not separated into blocks and is sent in one time
		piece.SetData(block)
		connInfo.UpdateNbAnswer(true)
		err := fileLeecher.pieceComplete(piece, conn)
		return err
	}

	// else we receive the piece in blocks
	blockId := int((begin - int64(fileLeecher.pieceSize)*int64(pieceId))) / CHUNKSIZE
	if piece.AddBlockToPiece(blockId, block) { // if we added a block we didn't have
		connInfo.UpdateNbAnswer(false)
		if piece.IsComplete() { // we have all the blocks of the piece
			err := fileLeecher.pieceComplete(piece, conn)
			return err
		} else { // we need to download the next blocks if possible
			piece.downloadNextBlocks()
		}
	}
	fileLeecher.m.Unlock()
	return nil
}

func (fileLeecher *FileLeecher) pieceComplete(piece *Piece, conn ShosetConn) error {

	pieceId := piece.GetId()

	ok := piece.CheckHash()
	if !ok {
		// download it again
		fileLeecher.requestedPieces[pieceId] = false
		delete(fileLeecher.piecesBeingRequested, pieceId)
		delete(fileLeecher.piecesToRequestPriority, pieceId) // in case if also there
		connInfo2 := fileLeecher.pieceAskedToConn[pieceId]
		connInfo2.RemovePieceRequested(pieceId)
		fileLeecher.m.Unlock()
		return fmt.Errorf("piece %d hash is not correct", pieceId)
	}
	fileLeecher.receivedPieces[pieceId] = true
	fileLeecher.requestedPieces[pieceId] = false
	delete(fileLeecher.piecesBeingRequested, pieceId)
	delete(fileLeecher.piecesToRequestPriority, pieceId) // in case if also there
	fileLeecher.nbPiecesReceived++
	conn.GetLogger().Info().Msgf("%v pieceComplete %d from %v total: %d / %d", conn.GetLocalAddress(), piece.GetId(), conn.GetRemoteAddress(), fileLeecher.nbPiecesReceived, fileLeecher.nbPieces)

	// write the piece to the copy file
	err := piece.WritePiece()
	if err != nil {
		fileLeecher.m.Unlock()
		return err
	}
	// update the connection info
	connInfo2 := fileLeecher.pieceAskedToConn[pieceId]
	connInfo2.RemovePieceRequested(pieceId)

	fileLeecher.SendHaveMessage(pieceId)
	// launch the download of the next pieces
	if !fileLeecher.isComplete() {
		fileLeecher.m.Unlock()
		return fileLeecher.DownloadNextPieces(conn)
	} else {
		// the unlock is in the enddownload function
		return fileLeecher.EndDownload()
	}
}

// usefull for the sending to know if we have the piece
func (fileLeecher *FileLeecher) haveChunk(begin int64, size int) bool {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()

	pieceId := int(begin / int64(fileLeecher.pieceSize))
	if size <= fileLeecher.pieceSize {
		return fileLeecher.receivedPieces[pieceId]
	}
	return false
}

// select a new piece to download with a specific algorithm
func (fileLeecher *FileLeecher) selectNewPiece(conn ShosetConn, connInfo ConnInfo) *Piece {
	pieceId := -1

	// we select first the pieces that are in the priority list
	for i := range fileLeecher.piecesToRequestPriority {
		if !fileLeecher.receivedPieces[i] && connInfo.HavePiece(i) { // if the conn has it
			delete(fileLeecher.piecesToRequestPriority, i)
			piece, ok := fileLeecher.piecesBeingRequested[i]
			if !ok {
				break
			}
			piece.SetConnInfo(connInfo)
			return piece
		}
	}
	// we chose the rarest piece among the available pieces
	rarestList := []int{}
	hightestRarity := -1
	for i, piece := range fileLeecher.receivedPieces {
		if !piece && !fileLeecher.requestedPieces[i] && connInfo.HavePiece(i) { // if we don't have it and we haven't requested it elsewhere and the peer has it
			newRarity := fileLeecher.pieceRarity[i]
			if hightestRarity == -1 || newRarity < hightestRarity {
				rarestList = []int{i}
				hightestRarity = newRarity
			} else if newRarity == hightestRarity {
				rarestList = append(rarestList, i)
			}
		}
	}
	if len(rarestList) != 0 {
		id := rand.Intn(len(rarestList))
		pieceId = rarestList[id]
		piece := NewPiece(fileLeecher.SyncFile, pieceId, fileLeecher.pieceSize, fileLeecher.hashMap[pieceId], connInfo)
		return piece
	}
	return nil
}

// new connection : we can ask him some pieces
func (fileLeecher *FileLeecher) launchDownload(conn ShosetConn) {
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if ok { // if we have already registered the conn
		if connInfo.IsReady() { // if we have already received the bitfield from the conn
			return
		} else { //  we ask the bitfield to the conn
			go fileLeecher.AskBitfieldFile(conn)
		}
	} else { // if we have not registered the conn
		conn.GetLogger().Info().Msg(conn.GetLocalAddress() + " launchDownload for " + conn.GetRemoteAddress() + " as a new conn")
		connInfo = NewConnInfo(conn, fileLeecher, fileLeecher.nbBlocks)
		fileLeecher.ConnInfoMap[conn] = connInfo
		fileLeecher.FileTransfer.AddFileLeecherToConn(fileLeecher, conn)
		go fileLeecher.AskBitfieldFile(conn)
		fileLeecher.SendInterested(conn, true)
	}
}

// send to a node that we are interested (or not) in the file
func (fileLeecher *FileLeecher) SendInterested(conn ShosetConn, interested bool) {
	message := msg.FileMessage{
		FileUUID: fileLeecher.SyncFile.GetUUID(),
		FileName: fileLeecher.File.GetName(),
		FileHash: fileLeecher.File.GetHash(),
	}
	message.InitMessageBase()
	if interested {
		message.MessageName = "interested"
	} else {
		message.MessageName = "notInterested"
	}
	fileLeecher.FileTransfer.SendMessage(message, conn)
}

// initialize the download of the file with a conn
func (fileLeecher *FileLeecher) InitDownload(conn ShosetConn) {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	fileLeecher.launchDownload(conn)
}

// we are asked to reduce our number of requests through this conn
func (fileLeecher *FileLeecher) ReduceRequests(conn ShosetConn) {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	conn.GetLogger().Info().Msg("-----------------------ReduceRequests from " + conn.GetRemoteAddress() + " to " + conn.GetLocalAddress())
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if ok {
		connInfo.DecreaseLevel()
	}
}

// we are not authorized to send requests to this conn (we retry after a timeout)
func (fileLeecher *FileLeecher) ReceiveUnauthorisedMessage(conn ShosetConn) {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	fileLeecher.removeConn(conn)
}

func (fileLeecher *FileLeecher) GetNbRequests(conn ShosetConn) int {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	connInfo, ok := fileLeecher.ConnInfoMap[conn]
	if ok {
		b, nbMax := connInfo.GetInfoRate()
		if b {
			return nbMax + 1000 // when asking for entire pieces : count more
		}
		return nbMax
	}
	return 0
}

// used to ask for more data of the file
func (fileLeecher *FileLeecher) DownloadNextPieces(conn ShosetConn) error {
	fileLeecher.m.Lock()
	connInfo, ok := fileLeecher.ConnInfoMap[conn]

	canAsk := true
	// we download in priority from files where we download block by block
	for _, pieceId2 := range connInfo.GetPiecesRequested() {
		if priority, ok := fileLeecher.piecesToRequestPriority[pieceId2]; ok && priority {
			// we don't want to continue downloading pieces in the priority list
			continue
		}
		piece2, ok := fileLeecher.piecesBeingRequested[pieceId2]
		if !ok {
			return fmt.Errorf("piece %d is not in the map", pieceId2)
		}
		if !piece2.GetNoBlock() {
			canAsk = piece2.downloadNextBlocks()
		}
	}

	fileLeecher.m.Unlock()
	if !canAsk { // we reached our limit of requests
		return nil
	}

	if ok && connInfo.IsReady() {
		entirePieces, _ := connInfo.GetInfoRate()
		if entirePieces { // we ask pieces by pieces
			count := 0 // we don't want to double the rate instantly, we go progressively
			for connInfo.CanRequestPiece() && count < 2 {
				fileLeecher.m.Lock()
				piece := fileLeecher.selectNewPiece(conn, connInfo)
				if piece != nil {
					pieceAdded := connInfo.AddPieceRequested(piece.pieceId)
					if pieceAdded {
						pieceId := piece.GetId()
						fileLeecher.piecesBeingRequested[pieceId] = piece
						fileLeecher.requestedPieces[pieceId] = true
						fileLeecher.pieceAskedToConn[pieceId] = connInfo
						piece.SetNoBlock(true)
						piece.downloadPiece()
						//fmt.Println(conn.GetLocalAddress(), "is downloading from", conn.GetRemoteAddress(), "piece", piece.GetId(), "with entire pieces : ", entirePieces, "and nb requests : ", nbMax)
					}
					fileLeecher.m.Unlock()
				} else {
					fileLeecher.m.Unlock()
					//fmt.Println("no piece to download from", conn.GetRemoteAddress())
					return nil
				}
				count++
			}

		} else if connInfo.GetNbRequestedPieces() <= 1 { // we ask blocks by blocks
			// we assume that everytime when we downloadNextBlocks, we fill all the requests possible. If there is a blank, it means we have requested all the blocks of the piece. We can search for another piece.
			// this way is smooth : we don't wait for the other piece to be complete.
			// we can download up to 2 pieces at the same time
			fileLeecher.m.Lock()
			piece := fileLeecher.selectNewPiece(conn, connInfo)
			if piece != nil {
				pieceAdded := connInfo.AddPieceRequested(piece.pieceId)
				if pieceAdded {
					pieceId := piece.GetId()
					fileLeecher.piecesBeingRequested[pieceId] = piece
					fileLeecher.pieceAskedToConn[pieceId] = connInfo
					fileLeecher.requestedPieces[pieceId] = true
					piece.SetNoBlock(false)
					//fmt.Println(conn.GetLocalAddress(), "trying to download next blocks from",piece.GetId())
					piece.downloadNextBlocks()
					//fmt.Println(conn.GetLocalAddress(), "is downloading from", conn.GetRemoteAddress(), "piece", piece.GetId(), "with entire pieces : ", entirePieces, "and nb requests : ", nbMax)
				}
				fileLeecher.m.Unlock()
			} else {
				fileLeecher.m.Unlock()
				//fmt.Println(conn.GetLocalAddress(), ": no piece to download from", conn.GetRemoteAddress())
				return nil
			}
		}
	}
	return nil
}

// when the download is complete, we can stop the leecher
func (fileLeecher *FileLeecher) EndDownload() error {
	delta := time.Now().UnixMilli() - fileLeecher.startTimestamp
	fileLeecher.FileTransfer.GetLogger().Info().Msg("-------------------------file " + fileLeecher.File.GetName() + " is complete ! in " + strconv.FormatInt(delta, 10) + " ms-------------------------")

	if fileLeecher.stopped {
		fileLeecher.m.Unlock()
		return fmt.Errorf("file leecher already stopped")
	}
	fileLeecher.stopped = true
	fileLeecher.m.Unlock()

	err := fileLeecher.SyncFile.EndDownload()
	if err != nil {
		hashMap := fileLeecher.SyncFile.GetCopyFile().GetHashMap()
		idList := make([]int, 0)
		for i := 0; i < len(hashMap); i++ {
			if hashMap[i] != fileLeecher.SyncFile.GetCopyFile().GetHash() {
				idList = append(idList, i)
			}
		}
		if len(idList) > 0 {
			fileLeecher.m.Lock()
			for _, id := range idList {
				fileLeecher.receivedPieces[id] = false

			}
			fileLeecher.stopped = false
			fileLeecher.m.Unlock()
			for conn := range fileLeecher.ConnInfoMap {
				fileLeecher.DownloadNextPieces(conn)
			}
		}
		return err
	}
	//fileLeecher.FileTransfer.FileSeeders.Store(fileLeecher.SyncFile.GetUUID(), &fileLeecher.FileSeeder)
	fileLeecher.FileTransfer.RemoveFileLeecher(fileLeecher)

	message := msg.FileMessage{
		MessageName: "downloadFinished",
		FileUUID:    fileLeecher.SyncFile.GetUUID(),
		FileName:    fileLeecher.File.GetName(),
		FileHash:    fileLeecher.File.GetHash(),
		FileVersion: fileLeecher.File.GetVersion(),
	}
	fileLeecher.FileTransfer.UserPush(&message)

	fileLeecher.FileTransfer.WriteRecords()
	return err
}

func (fileLeecher *FileLeecher) IsComplete() bool {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	return fileLeecher.isComplete()
}
func (fileLeecher *FileLeecher) isComplete() bool {
	return fileLeecher.nbPiecesReceived == fileLeecher.nbPieces
}

func (fileLeecher *FileLeecher) GetBitfield() []bool {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	bitfield := make([]bool, fileLeecher.nbPieces)
	for i := 0; i < fileLeecher.nbPieces; i++ {
		bitfield[i] = fileLeecher.receivedPieces[i]
	}
	return bitfield
}

func (fileLeecher *FileLeecher) GetPieceSize() int {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	return fileLeecher.pieceSize
}

func (fileLeecher *FileLeecher) GetConnInfo() map[ShosetConn]ConnInfo {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	newMap := make(map[ShosetConn]ConnInfo)
	for conn, connInfo := range fileLeecher.ConnInfoMap {
		newMap[conn] = connInfo
	}
	return newMap
}

// fill the map with the sum of the rate of each conn and the number of rate we add
// it will be used in FileTransfer to monitor the rate of each conn
func (fileLeecher *FileLeecher) UpdateGlobalRate(connMapSum map[ShosetConn]int, connMapCount map[ShosetConn]int) {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	for conn, connInfo := range fileLeecher.ConnInfoMap {
		sum, ok := connMapSum[conn]
		if !ok {
			connMapSum[conn] = connInfo.GetRate()
			connMapCount[conn] = 1
		} else {
			connMapSum[conn] = sum + connInfo.GetRate()
			connMapCount[conn] += 1
		}
	}
}

// ask the bitfield of the file to the conn
// to execute in a goroutine
func (fileLeecher *FileLeecher) AskBitfieldFile(conn ShosetConn) {
	count := 0
	for count < 3 { // we try 3 times max
		fileLeecher.m.Lock()
		_, ok := fileLeecher.GetBitfiledFile[conn]
		if ok { // we already have a goroutine trying to ask for the bitfield
			fileLeecher.m.Unlock()
			return
		}
		// if we don't hav a goroutine asking for the bitfield, we create one
		fileLeecher.GetBitfiledFile[conn] = true
		connInfo, ok := fileLeecher.ConnInfoMap[conn]
		fileLeecher.m.Unlock()
		if ok {
			if connInfo.IsReady() { // we received the bitfield
				return
			} else {
				count++
				askBitfield := msg.FileMessage{
					MessageName: "askBitfield",
					FileUUID:    fileLeecher.SyncFile.GetUUID(),
					FileName:    fileLeecher.File.GetName(),
					FileHash:    fileLeecher.File.GetHash(),
					PieceSize:   fileLeecher.File.GetPieceSize(),
				}
				askBitfield.InitMessageBase()
				fileLeecher.FileTransfer.SendMessage(askBitfield, conn)
				time.Sleep(10 * time.Second)
			}
		}
	}
}

// check the timeout of each pieces we are requesting
// to execute regularly
func (fileLeecher *FileLeecher) CheckRequestTimeout() {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	for pieceId, piece := range fileLeecher.piecesBeingRequested {
		if !fileLeecher.piecesToRequestPriority[pieceId] { // if the piece is not a standalone piece
			fileLeecher.m.Unlock()
			timeoutOccured := piece.CheckRequestTimeout()
			fileLeecher.m.Lock()
			if timeoutOccured {
				connInfo := piece.ConnInformation
				if entire, nb := connInfo.GetInfoRate(); !entire && nb == 1 {
					// we remove the conn if there is a timeout and the rate is already really slow
					fileLeecher.removeConn(connInfo.GetConn())
				}
			}
		}
	}
}

// when a node has difficulty to give us the piece (too slow or disconnected), we place the piece in a priority list to download.
func (fileLeecher *FileLeecher) AddToRequestPriority(pieceId int, fromConnInfo ConnInfo) {
	fileLeecher.m.Lock()
	conn := fromConnInfo.GetConn()
	fromConnInfo.GetConn().GetLogger().Info().Msgf("%v : adding %d to priority", conn.GetLocalAddress(), pieceId)
	fileLeecher.piecesToRequestPriority[pieceId] = true
	fromConnInfo.RemovePieceRequested(pieceId)
	fileLeecher.m.Unlock()
	for c := range fileLeecher.ConnInfoMap {
		fileLeecher.DownloadNextPieces(c)
	}
}

func (fileLeecher *FileLeecher) GetHash() string {
	fileLeecher.m.Lock()
	defer fileLeecher.m.Unlock()
	return fileLeecher.hash
}
