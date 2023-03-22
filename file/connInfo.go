package fileSync

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

/*
connInfo stands for Connection Information. It contains all the information about a connection linked to a specific file.
The amount of requests and data we can ask at the same time is stored in it.
It is separated in levels.
*/

// ConnInfo ...
type ConnInfo interface {
	// decrease the level of the connection (the number of requests we can send to him at the same time)
	DecreaseLevel()
	// when we had a decrease, it wait a certain time before retrying to send requests to him
	ReAskAfterDecrease()
	// when we had a decrease, it wait a certain time before retrying to increase the level of the connection
	RetryToIncrease(currentIncreaseTime int64, currentIncreaseInc int)
	// piece = true if we download piece by piece
	UpdateNbAnswer(piece bool)
	GetInfoRate() (bool, int)
	// ad the number of bytes we received from this connection
	AddBytesSeries(bytes int)
	GetRate() int
	GetNbRequestedPieces() int
	GetAvailable() (bool, int, int64)
	// reset also the timestamp of the last time the conn was available
	SetAvailable(available bool)
	RemovePieceRequested(pieceId int)
	AddPieceRequested(pieceId int) bool
	AddRequestBlock() bool
	RemoveRequestBlock()
	CanRequestBlock() bool
	CanRequestPiece() bool
	HavePiece(pieceId int) bool
	// the node has a new piece. We add it to the bitfield
	AddPiece(pieceId int)
	IsReady() bool
	SetBitfield(bitfield []bool)
	GetIncreaseTime() int64
	String() string
	GetPiecesRequested() []int
	HadARecentDecrease() bool
	GetNbRequests() int
	GetConn() ShosetConn
}

// information about a connection
type ConnectionInformation struct {
	Conn               ShosetConn   // connection
	piecesRequested    map[int]bool // list of pieces we have requested to him
	bitfield           []bool       // bitfield of the pieces he has
	nbBlocks           int          // number of blocks in a piece
	askingEntirePieces bool         // true if we are asking diretly entire pieces to him (no blocks separation)
	nbMaxRequests      int          // number of requests can send to him at the same time
	nbRequests         int          // number of requests we have sent to him
	nbAnsweredRequests int          // number of requests which have been answered
	timestampSeries    int64        // timestamp of the first request to calculate the speed
	totalBytesSeries   int          // total bytes received in the series
	available          bool         // true if we are authorised to send requests to him
	availableTime      int64        // timestamp of the last time the conn was available
	availableInc       int          // this number grow higher if the peer is not available
	increase           bool         // true if we can increase the number of requests
	increaseTime       int64        // timestamp of the last time we increased the number of requests
	increaseInc        int          // this number grow higher if we can't increase the number of requests

	lastRequestTime  int64 // timestamp of the last request we have sent to him
	lastIncreaseTime int64 // timestamp of the last time we increased the number of requests

	m sync.RWMutex
}

func NewConnInfo(conn ShosetConn, blocks int) *ConnectionInformation {
	return &ConnectionInformation{
		Conn:            conn,
		nbMaxRequests:   4,
		nbBlocks:        blocks,
		piecesRequested: make(map[int]bool),
		increase:        true,
		available:       true,
		increaseInc:     100,
		availableInc:    100,
	}
}

// increase the lebel of the connection (the number of requests we can send to him at the same time)
func (connInfo *ConnectionInformation) increaseLevel() {
	if connInfo.increase && time.Now().UnixMilli()-connInfo.lastIncreaseTime > 3000 { // we don't want to increase too fast
		if connInfo.askingEntirePieces {
			connInfo.nbMaxRequests = Min(connInfo.nbMaxRequests*2, 16) // max 16 requests
		} else {
			if connInfo.nbMaxRequests >= connInfo.nbBlocks { // if we reached the limit of blocks
				connInfo.askingEntirePieces = true
				connInfo.nbMaxRequests = 1
			} else {
				connInfo.nbMaxRequests = connInfo.nbMaxRequests * 2
			}
		}
		connInfo.nbAnsweredRequests = 0
		connInfo.GetConn().GetLogger().Info().Msgf("increase level to : download piece by piece : %t with max %d requests.", connInfo.askingEntirePieces, connInfo.nbMaxRequests)
		connInfo.lastIncreaseTime = time.Now().UnixMilli()
	}
}

// decrease the level of the connection (the number of requests we can send to him at the same time)
func (connInfo *ConnectionInformation) DecreaseLevel() {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	// we don't want to decrease too fast
	if time.Now().UnixMilli()-connInfo.increaseTime > TIMEBEFOREANOTHERDECREASE && time.Now().UnixMilli()-connInfo.lastRequestTime < TIMEBEFOREANOTHERDECREASE { // if the inc have not been increased for the last ...s
		conn := connInfo.Conn
		connInfo.GetConn().GetLogger().Info().Msgf("from %v to %v decrease level from : download piece by piece : %t with max %d requests.", conn.GetLocalAddress(), conn.GetRemoteAddress(), connInfo.askingEntirePieces, connInfo.nbMaxRequests)
		connInfo.increase = false
		if connInfo.askingEntirePieces {
			if connInfo.nbMaxRequests == 1 { // if we reached the limit of pieces
				connInfo.askingEntirePieces = false
				connInfo.nbMaxRequests = connInfo.nbBlocks / 2
			} else {
				connInfo.nbMaxRequests = connInfo.nbMaxRequests / 2
			}
		} else {
			connInfo.nbMaxRequests = Max(connInfo.nbMaxRequests/2, 2)
		}
		connInfo.nbAnsweredRequests = 0

		connInfo.increaseInc = Min(10000, connInfo.increaseInc*10) // max 10s x 10
		connInfo.increaseTime = time.Now().UnixMilli()

		// we authorise to increase the level of the connection after a certain time
		go connInfo.RetryToIncrease(connInfo.increaseTime, connInfo.increaseInc)

		if connInfo.available { // if the conn is available
			connInfo.available = false
			// we authorise to send requests to him after a certain time
			go connInfo.ReAskAfterDecrease()
		}
	}
}

// when we had a decrease, it wait a certain time before retrying to send requests to him
func (connInfo *ConnectionInformation) ReAskAfterDecrease() {
	connInfo.m.Lock()
	conn := connInfo.Conn
	connInfo.GetConn().GetLogger().Info().Msg(conn.GetLocalAddress() + " is sleeping 3 seconds before downloading again from " + conn.GetRemoteAddress())
	time.Sleep(time.Duration(TIMENOTASKINGDURINGDECREASE) * time.Millisecond)
	connInfo.m.Unlock()
	connInfo.SetAvailable(true)
	connInfo.m.Lock()
	connInfo.GetConn().GetLogger().Info().Msgf(conn.GetLocalAddress()+": sleeping finished, downloading again from : download piece by piece : %t with max %d requests.", connInfo.askingEntirePieces, connInfo.nbMaxRequests)
	connInfo.m.Unlock()
	// TODO : continue downloading
}

// when we had a decrease, it wait a certain time before retrying to increase the level of the connection
func (connInfo *ConnectionInformation) RetryToIncrease(currentIncreaseTime int64, currentIncreaseInc int) {
	time.Sleep(time.Duration(currentIncreaseInc*(rand.Intn(10)+1)) * time.Millisecond) // we sleep the number of time we should
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	if currentIncreaseTime == connInfo.increaseTime { // if we didn't received another decrease order while we were sleeping
		connInfo.increase = true
		go connInfo.resetIncreaseInc()
	}
}

func (connInfo *ConnectionInformation) resetIncreaseInc() {
	notBreaking := true
	for notBreaking {
		connInfo.m.Lock()
		if time.Now().UnixMilli()-connInfo.increaseTime > 6000 { // if we can still increase : everything back to normal : we reset the increase inc
			connInfo.increaseInc = Max(100, connInfo.increaseInc/10)
		} else {
			notBreaking = false
		}
		connInfo.m.Unlock()
		time.Sleep(5 * time.Second)
	}
}

// piece = true if we download piece by piece
func (connInfo *ConnectionInformation) UpdateNbAnswer(piece bool) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	if (connInfo.askingEntirePieces && piece) || (!connInfo.askingEntirePieces && !piece) {
		connInfo.nbAnsweredRequests++
		connInfo.nbRequests = Max(connInfo.nbRequests-1, 0)
	} else { // if the level has changed recently, we don't want to increase the number of answered requests
		connInfo.nbRequests = Max(connInfo.nbRequests-1, 0)
	}
	if connInfo.nbAnsweredRequests >= connInfo.nbMaxRequests {
		connInfo.increaseLevel()
	}
}

func (connInfo *ConnectionInformation) GetInfoRate() (bool, int) {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.askingEntirePieces, connInfo.nbMaxRequests
}

// ad the number of bytes we received from this connection
func (connInfo *ConnectionInformation) AddBytesSeries(bytes int) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	connInfo.totalBytesSeries += bytes
}

func (connInfo *ConnectionInformation) resetRate() {
	connInfo.timestampSeries = time.Now().UnixMilli()
	connInfo.totalBytesSeries = 0
}

func (connInfo *ConnectionInformation) GetRate() int {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	deltaTime := Max64(time.Now().UnixMilli()-connInfo.timestampSeries, 1) // to not divide by 0
	rate := int(int64(connInfo.totalBytesSeries*1000) / deltaTime)         // bytes per second
	connInfo.resetRate()
	return rate
}

func (connInfo *ConnectionInformation) GetNbRequestedPieces() int {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return len(connInfo.piecesRequested)
}

func (connInfo *ConnectionInformation) GetAvailable() (bool, int, int64) {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.available, connInfo.availableInc, connInfo.availableTime
}

// reset also the timestamp of the last time the conn was available
func (connInfo *ConnectionInformation) SetAvailable(available bool) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	connInfo.available = available
	if !available {
		connInfo.availableTime = 0
		connInfo.availableInc = Min(100000, connInfo.availableInc*10)
	} else {
		connInfo.availableInc = 1
		connInfo.availableTime = time.Now().UnixMilli()
	}
}

func (connInfo *ConnectionInformation) RemovePieceRequested(pieceId int) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	delete(connInfo.piecesRequested, pieceId)
	connInfo.nbRequests = Max(connInfo.nbRequests-1, 0)
}

func (connInfo *ConnectionInformation) AddPieceRequested(pieceId int) bool {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	if connInfo.available {
		if connInfo.askingEntirePieces {
			if connInfo.nbRequests < connInfo.nbMaxRequests {
				connInfo.piecesRequested[pieceId] = true
				connInfo.lastRequestTime = time.Now().UnixMilli()
				connInfo.nbRequests++
				return true
			}
		} else {
			if connInfo.nbRequests < connInfo.nbMaxRequests && len(connInfo.piecesRequested) <= 1 {
				connInfo.piecesRequested[pieceId] = true
				return true
			}
		}
	}
	return false
}

func (connInfo *ConnectionInformation) AddRequestBlock() bool {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	if !connInfo.askingEntirePieces && connInfo.available && connInfo.nbRequests < connInfo.nbMaxRequests {
		connInfo.nbRequests++
		connInfo.lastRequestTime = time.Now().UnixMilli()
		return true
	}
	return false
}

func (connInfo *ConnectionInformation) RemoveRequestBlock() {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	connInfo.nbRequests = Max(connInfo.nbRequests-1, 0)
}

func (connInfo *ConnectionInformation) CanRequestBlock() bool {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.available && (connInfo.nbRequests < connInfo.nbMaxRequests || connInfo.askingEntirePieces)
}

func (connInfo *ConnectionInformation) CanRequestPiece() bool {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.available && connInfo.askingEntirePieces && connInfo.nbRequests < connInfo.nbMaxRequests
}

func (connInfo *ConnectionInformation) HavePiece(pieceId int) bool {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.bitfield[pieceId]
}

// the node has a new piece. We add it to the bitfield
func (connInfo *ConnectionInformation) AddPiece(pieceId int) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	connInfo.bitfield[pieceId] = true
}

func (connInfo *ConnectionInformation) IsReady() bool {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return len(connInfo.bitfield) > 0
}

func (connInfo *ConnectionInformation) SetBitfield(bitfield []bool) {
	connInfo.m.Lock()
	defer connInfo.m.Unlock()
	connInfo.bitfield = bitfield
}

func (connInfo *ConnectionInformation) GetIncreaseTime() int64 {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.increaseTime
}

func (connInfo *ConnectionInformation) String() string {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return fmt.Sprintf("askingEntirePieces : %t, nbMaxRequests : %v, nbRequests : %v, nbAnsweredRequests : %v, increase : %v, increaseTime : %v, increaseInc : %v, available : %v, availableTime : %v, availableInc : %v, piecesRequested : %v", connInfo.askingEntirePieces, connInfo.nbMaxRequests, connInfo.nbRequests, connInfo.nbAnsweredRequests, connInfo.increase, connInfo.increaseTime, connInfo.increaseInc, connInfo.available, connInfo.availableTime, connInfo.availableInc, connInfo.piecesRequested)
}

func (connInfo *ConnectionInformation) GetPiecesRequested() []int {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	pieces := []int{}
	connInfo.GetConn().GetLogger().Info().Msgf("requested", connInfo.piecesRequested)
	for pieceId := range connInfo.piecesRequested {
		pieces = append(pieces, pieceId)
	}
	return pieces
}

func (connInfo *ConnectionInformation) HadARecentDecrease() bool {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	// true if the last decrease was less than 3 seconds ago or if the number of requests is less than half of the max number of requests
	return time.Now().UnixMilli()-connInfo.lastRequestTime > 3000
}

func (connInfo *ConnectionInformation) GetNbRequests() int {
	connInfo.m.RLock()
	defer connInfo.m.RUnlock()
	return connInfo.nbRequests
}

func (connInfo *ConnectionInformation) GetConn() ShosetConn {
	return connInfo.Conn
}
