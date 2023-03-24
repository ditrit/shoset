package fileSync

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog"
)

/*
This file handle messages for file transfer.
It monitor the flow rate of the network and send congestion messages to the network.

*/

type Rate struct {
	minTimeConn map[ShosetConn]int // min RTT for each conn

	m sync.RWMutex
}

func NewRate() *Rate {
	var Rate Rate
	Rate.minTimeConn = make(map[ShosetConn]int)
	return &Rate
}

func (rate *Rate) GetMin(conn ShosetConn) int {
	rate.m.Lock()
	defer rate.m.Unlock()
	min, ok := rate.minTimeConn[conn]
	if !ok {
		return 10000
	}
	return min
}

func (rate *Rate) GetNbConn() int {
	rate.m.Lock()
	defer rate.m.Unlock()
	return len(rate.minTimeConn)
}

func (rate *Rate) SetMin(conn ShosetConn, min int) {
	rate.m.Lock()
	defer rate.m.Unlock()
	rate.minTimeConn[conn] = min
}

type FileTransfer interface {
	Init(library FileLibrary, logger zerolog.Logger, userMessageQueue *msg.Queue, broadcast func(message *msg.FileMessage))
	RemoveFileSeeder(fileSeeder *FileSeeder)
	AddFileLeecher(fileLeecher *FileLeecher)
	AddFileLeecherToConn(fileLeecher *FileLeecher, conn ShosetConn)
	RemoveFileLeecher(fileLeecher *FileLeecher)
	ReceiveCongestionMessage(conn ShosetConn)
	SendCongestionMessage(conn ShosetConn)
	ReceiveUnauthorisedMessage(conn ShosetConn)
	SendUnauthorisedMessage(conn ShosetConn)
	// receive a message "authorised" from a conn
	ReceiveAuthorisedMessage(conn ShosetConn)
	// send a the message "authorised" to a conn
	SendAuthorisedMessage(conn ShosetConn)
	ReceiveHaveMessage(conn ShosetConn, message *msg.FileMessage)
	ReceiveBitfieldMessage(conn ShosetConn, message *msg.FileMessage)
	// if someone is asking us our bitfield of a file
	ReceiveAskBitfieldMessage(conn ShosetConn, message *msg.FileMessage)
	ReceiveInterestedMessage(conn ShosetConn, message *msg.FileMessage)
	IsAuthorised(conn ShosetConn) bool
	// a new conn is sending us requests
	AddUploadConn(conn ShosetConn)
	// remove a conn from the list of conn that are downloading from us
	RemoveConn(conn ShosetConn)
	AskChunk(conn ShosetConn, message *msg.FileMessage, sendRate bool)
	ReceiveChunk(conn ShosetConn, message *msg.FileMessage) error
	ReceiveAskChunk(conn ShosetConn, message *msg.FileMessage) error
	ReceiveAskInfoMessage(conn ShosetConn, message *msg.FileMessage)
	GetMissingLength() int64
	DecreaseMissingLength(length int)
	// keep asking info if we don't have this file
	// to launch in a goroutine
	AskInfoFile(conn ShosetConn, fileState FileState)
	InitLeecher(syncFile SyncFile, conn ShosetConn)
	ReceiveMessage(message *msg.FileMessage, conn ShosetConn)
	HandleReceiveMessage(messageQueue *MessageQueue, conn ShosetConn)
	HandleReceiveMessageFromQueue(fileMessage msg.FileMessage, c ShosetConn) error
	SendMessage(message msg.FileMessage, conn ShosetConn)
	HandleSendMessage(messageQueue *MessageQueue, conn ShosetConn)
	HandleSendMessageFromQueue(fileMessage msg.FileMessage, c ShosetConn) error
	SetExternalCommands(ec *ExternalCommands)
	SetNbConn(nbConn int)
	GetLibrary() FileLibrary
	Broadcast(message *msg.FileMessage)
	UserPush(message *msg.FileMessage)
	DeleteLeecher(fileUUID string)
	GetLeecher(fileUUID string) *FileLeecher
	WriteRecords()
	GetLogger() *zerolog.Logger
	GetReceiveQueue(conn *ShosetConn) *MessageQueue
}

// global struct to handle multiple file transfer
type FileTransferImpl struct {
	// attributes with no mutex to be accessible quickly
	Library             FileLibrary // file library
	FileSeeders         sync.Map    //map[string]*FileSeeder  // map of fileSeeder
	FileLeechers        sync.Map    //map[string]*FileLeecher // map of fileLeecher
	FileLeechersPerConn sync.Map    //map[ShosetConn][]*FileLeecher // map of fileLeecher per conn
	AuthorisedConn      sync.Map    //map[ShosetConn]bool     // list of authorised conn : conn that can send us requests
	ConnLastTime        sync.Map    //map[ShosetConn]int64 // last time a conn sent a request
	MissingDataConn     sync.Map    //map[ShosetConn]int64 // map of missing data (in bytes) per conn
	DecreaseConn        sync.Map    //map[ShosetConn]int64 // map of conn with a timestamp to know the last time we send a congestion message
	RTTRate             *Rate

	missingLength    int64             // length of our missing data
	GetInfoMap       map[string]string // map of {uuid : hash} hash being asked
	stop             chan bool
	Logger           zerolog.Logger
	SendQueue        sync.Map // map[*ShosetConn]MessageQueue
	ReceiveQueue     sync.Map // map[*ShosetConn]MessageQueue
	LogRecords       LogRateTransfer
	broadcast        func(message *msg.FileMessage)
	userMessageQueue *msg.Queue
	nbConn           int

	ExternalCommands *ExternalCommands
	m                sync.Mutex
}

func (ft *FileTransferImpl) Init(library FileLibrary, logger zerolog.Logger, userMessageQueue *msg.Queue, broadcast func(message *msg.FileMessage)) {
	ft.FileSeeders = sync.Map{}
	ft.FileLeechers = sync.Map{}
	ft.FileLeechersPerConn = sync.Map{}
	ft.AuthorisedConn = sync.Map{}
	ft.ConnLastTime = sync.Map{}
	ft.MissingDataConn = sync.Map{}
	ft.DecreaseConn = sync.Map{}
	ft.RTTRate = NewRate()
	ft.broadcast = broadcast

	ft.GetInfoMap = make(map[string]string)
	ft.stop = make(chan bool)
	ft.Logger = logger
	ft.userMessageQueue = userMessageQueue
	ft.SendQueue = sync.Map{}
	ft.ReceiveQueue = sync.Map{}

	ft.watchRequestsTimeout(1000)
	ft.watchRTT(500)

	ft.Library = library
	library.SetFileTransfer(ft)
	ft.LogRecords = *NewLogRateTransfer(filepath.Join(ft.Library.GetDir(), PATH_COPY_FILES, ".info", "flowrate.csv"))

	ft.LogRecords.createMonitorFile()
}

func (ft *FileTransferImpl) RemoveFileSeeder(fileSeeder *FileSeeder) {
	ft.FileSeeders.Delete(fileSeeder.SyncFile.GetUUID())
}

func (ft *FileTransferImpl) AddFileLeecher(fileLeecher *FileLeecher) {
	ft.FileLeechers.Store(fileLeecher.SyncFile.GetUUID(), fileLeecher)
	ft.m.Lock()
	delete(ft.GetInfoMap, fileLeecher.SyncFile.GetUUID())
	ft.missingLength += fileLeecher.File.GetSize()
	ft.m.Unlock()
}

func (ft *FileTransferImpl) AddFileLeecherToConn(fileLeecher *FileLeecher, conn ShosetConn) {
	leechers, ok := ft.FileLeechersPerConn.Load(conn)
	if ok {
		ft.FileLeechersPerConn.Store(conn, append(leechers.([]*FileLeecher), fileLeecher))
	} else {
		ft.FileLeechersPerConn.Store(conn, []*FileLeecher{fileLeecher})
	}
}

func (ft *FileTransferImpl) RemoveFileLeecher(fileLeecher *FileLeecher) {
	ft.m.Lock()
	defer ft.m.Unlock()
	uuid := fileLeecher.SyncFile.GetUUID()
	ft.FileLeechers.Delete(uuid)
	connInfo := fileLeecher.GetConnInfo()
	for conn := range connInfo {
		leechersList, ok := ft.FileLeechersPerConn.Load(conn)
		if ok {
			leechers := leechersList.([]*FileLeecher)
			for i, leecher := range leechers {
				if leecher.SyncFile.GetUUID() == uuid {
					leechers = append(leechers[:i], leechers[i+1:]...)
					break
				}
			}
			if len(leechers) == 0 {
				ft.FileLeechersPerConn.Delete(conn)
				fileLeecher.SendInterested(conn, false)
			} else {
				ft.FileLeechersPerConn.Store(conn, leechers)
			}
		}
	}
}

func (ft *FileTransferImpl) ReceiveCongestionMessage(conn ShosetConn) {
	ft.reduceRequests(conn)
}

func (ft *FileTransferImpl) SendCongestionMessage(conn ShosetConn) {
	message := msg.FileMessage{
		MessageName: "congestion",
	}
	message.InitMessageBase()
	ft.SendMessage(message, conn)
}

func (ft *FileTransferImpl) ReceiveUnauthorisedMessage(conn ShosetConn) {
	leechers, ok := ft.FileLeechersPerConn.Load(conn)
	if ok {
		for _, leecher := range leechers.([]*FileLeecher) {
			leecher.ReceiveUnauthorisedMessage(conn)
		}
	}
}

func (ft *FileTransferImpl) SendUnauthorisedMessage(conn ShosetConn) {
	message := msg.FileMessage{
		MessageName: "unauthorised",
	}
	message.InitMessageBase()
	ft.SendMessage(message, conn)
	conn.GetLogger().Info().Msg("------------------------------" + conn.GetRemoteAddress() + "is unauthorised -------------------------------")
}

// receive a message "authorised" from a conn
func (ft *FileTransferImpl) ReceiveAuthorisedMessage(conn ShosetConn) {
	ft.AuthorisedConn.Store(conn, true)
}

// send a the message "authorised" to a conn
func (ft *FileTransferImpl) SendAuthorisedMessage(conn ShosetConn) {
	message := msg.FileMessage{
		MessageName: "authorised",
	}
	message.InitMessageBase()
	ft.SendMessage(message, conn)
}

func (ft *FileTransferImpl) ReceiveHaveMessage(conn ShosetConn, message *msg.FileMessage) {
	leecher, ok := ft.FileLeechers.Load(message.FileUUID)
	if ok {
		leecher.(*FileLeecher).ReceiveHaveMessage(conn, message.PieceNumber)
	}
}

func (ft *FileTransferImpl) ReceiveBitfieldMessage(conn ShosetConn, message *msg.FileMessage) {
	leecher, ok := ft.FileLeechers.Load(message.FileUUID)
	if ok {
		leecher.(*FileLeecher).ReceiveBitfieldMessage(conn, message.Bitfield)
	} else {
		log.Println("ReceiveBitfieldMessage: leecher not found for uuid :", message.FileUUID)
	}
}

// if someone is asking us our bitfield of a file
func (ft *FileTransferImpl) ReceiveAskBitfieldMessage(conn ShosetConn, message *msg.FileMessage) {
	answer := msg.FileMessage{
		MessageName: "sendBitfield",
		FileUUID:    message.FileUUID,
		FileName:    message.FileName,
		FileHash:    message.FileHash,
		PieceSize:   message.PieceSize,
	}
	answer.InitMessageBase()
	syncFile, err := ft.Library.GetFile(message.FileUUID)
	if err != nil { // we don't have this file and we are not downloading it
		// we try to have the leecher if we are downloading it.
		ft.SendMessage(answer, conn)
	} else { // we have this file or we are downloading it
		if syncFile.GetCopyFile().GetHash() != message.FileHash { // we have this file but it is not the same
			// we send short info about the file
			ft.SendMessage(answer, conn)
			shortInfo := syncFile.GetShortInfoMsg()
			ft.SendMessage(shortInfo, conn)
		} else { // we have this file and it is the same
			leecher, ok := ft.FileLeechers.Load(message.FileUUID)
			if ok { // we have a leecher for this file
				leecher := leecher.(*FileLeecher)
				bitfield := leecher.GetBitfield()
				answer.Bitfield = bitfield
				ft.SendMessage(answer, conn)
				// we add this connection too because he is probably going to download the file too
				leecher.InitDownload(conn)
			} else { // we don't have a leecher for this file
				nbPieces := int(syncFile.GetCopyFile().GetSize() / int64(message.PieceSize))
				bitfield := make([]bool, nbPieces)
				for i := 0; i < nbPieces; i++ {
					bitfield[i] = true
				}
				answer.Bitfield = bitfield
				ft.SendMessage(answer, conn)
			}
		}
	}
}

func (ft *FileTransferImpl) ReceiveInterestedMessage(conn ShosetConn, message *msg.FileMessage) {
	fileUUID := message.FileUUID
	seeder, ok := ft.FileSeeders.Load(fileUUID)
	if ok {
		if message.MessageName == "interested" {
			conn.GetLogger().Info().Msg("-------------" + conn.GetRemoteAddress() + "is interested in" + conn.GetLocalAddress())
			seeder.(*FileSeeder).InterestedConn.Store(conn, true)
		} else {
			seeder.(*FileSeeder).InterestedConn.Delete(conn)
			conn.GetLogger().Info().Msg("-------------" + conn.GetRemoteAddress() + "is really not interested in" + conn.GetLocalAddress())

			count := 0
			seeder.(*FileSeeder).InterestedConn.Range(func(key, value interface{}) bool {
				count++
				return true
			})
			if count == 0 {
				ft.LogRecords.writeRecords()
			}
		}
		return
	}
	leecher, ok := ft.FileLeechers.Load(fileUUID)
	if ok {
		if message.MessageName == "interested" {
			leecher.(*FileLeecher).InterestedConn.Store(conn, true)
			conn.GetLogger().Info().Msg("-------------" + conn.GetRemoteAddress() + "is interested in" + conn.GetLocalAddress())
		} else {
			leecher.(*FileLeecher).InterestedConn.Delete(conn)
			conn.GetLogger().Info().Msg("-------------" + conn.GetRemoteAddress() + "is really not interested in" + conn.GetLocalAddress())

			count := 0
			leecher.(*FileLeecher).InterestedConn.Range(func(key, value interface{}) bool {
				count++
				return true
			})
			if count == 0 {
				ft.LogRecords.writeRecords()
			}
		}
		return
	}
}

func (ft *FileTransferImpl) IsAuthorised(conn ShosetConn) bool {
	_, ok := ft.AuthorisedConn.Load(conn)
	return ok
}

// a new conn is sending us requests
func (ft *FileTransferImpl) AddUploadConn(conn ShosetConn) {
	ft.AuthorisedConn.Store(conn, true)
	ft.ConnLastTime.Store(conn, time.Now().Unix())
}

// remove a conn from the list of conn that are downloading from us
func (ft *FileTransferImpl) RemoveConn(conn ShosetConn) {
	ft.AuthorisedConn.Delete(conn)
	ft.ConnLastTime.Delete(conn)
}

func (ft *FileTransferImpl) monitorDecrease(decreasingConn []ShosetConn) {
	mapCongestionConn := make(map[ShosetConn]bool)
	mapPiecesConn := make(map[ShosetConn]int64)

	// get the ordered list of conn (ascending order)
	ft.MissingDataConn.Range(func(key, value interface{}) bool {
		mapPiecesConn[key.(ShosetConn)] = value.(int64)
		return true
	})
	connList := make([]ShosetConn, 0, ft.RTTRate.GetNbConn()) // list of all the conn we are uploading to
	ft.AuthorisedConn.Range(func(key, value interface{}) bool {
		connList = append(connList, key.(ShosetConn))
		return true
	})
	sort.Slice(connList, func(i, j int) bool { return mapPiecesConn[connList[i]] < mapPiecesConn[connList[j]] })
	nbConn := len(connList)
	mapPlaceConn := make(map[ShosetConn]int)
	for i, conn := range connList {
		mapPlaceConn[conn] = i * 100 / nbConn
	}
	for _, conn := range decreasingConn {
		// we don't take care of the conn that have decreased for less than 2s
		isDecreasing, ok := ft.DecreaseConn.Load(conn)
		if ok {
			if time.Now().UnixMilli()-isDecreasing.(int64) < 4000 { // if the conn is decreasing for less than ...s
				//fmt.Println("conn", conn.GetRemoteAddress(), "is decreasing for less than 4s")
				continue
			} else {
				ft.DecreaseConn.Delete(conn)
			}
		}
		for i, c := range connList {
			rd := rand.Intn(100) + 1
			if i*100/nbConn*mapPlaceConn[conn]/100 > rd { // with random, higher chance to be picked if have a lot of pieces and the decreasing conn have not a lot of pieces
				mapCongestionConn[c] = true
			}
		}
	}
	for conn, b := range mapCongestionConn {
		if b {
			ft.DecreaseConn.Store(conn, time.Now().UnixMilli())
			ft.SendCongestionMessage(conn)
			conn.GetLogger().Info().Msg("-----------------------sending ReduceRequests from " + conn.GetLocalAddress() + " to " + conn.GetRemoteAddress())
		}

	}
	for _, conn := range decreasingConn {
		ft.reduceRequests(conn)
	}
}

func (ft *FileTransferImpl) reduceRequests(conn ShosetConn) {
	listFileLeecher, ok := ft.FileLeechersPerConn.Load(conn)
	if !ok {
		return
	}
	worseLeecher := listFileLeecher.([]*FileLeecher)[0]
	worseLeecherScore := 0
	for _, fileLeecher := range listFileLeecher.([]*FileLeecher) {
		score := fileLeecher.GetNbRequests(conn)
		if score > worseLeecherScore { // reduce the number of requests for the file that does the most requests
			worseLeecher = fileLeecher
			worseLeecherScore = score
		}
	}
	worseLeecher.ReduceRequests(conn)
}

func (ft *FileTransferImpl) updateRTT() {
	toDecrease := []ShosetConn{}
	ft.AuthorisedConn.Range(func(key, value interface{}) bool {
		conn := key.(ShosetConn)
		min := ft.RTTRate.GetMin(conn)
		tcp, err := conn.GetTCPInfo()
		if err != nil {
			fmt.Println(err)
		}
		rtt := int(tcp.Rtt) / 1000
		if rtt < min {
			ft.RTTRate.SetMin(conn, Max(rtt, 10))
		} else if rtt > min*8 {
			//fmt.Println(conn.GetLocalAddress(), "diminish", conn.GetRemoteAddress(), "because rtt is", rtt, "and min is", min)
			toDecrease = append(toDecrease, conn)
		}
		return true
	})
	if len(toDecrease) > 0 {
		ft.monitorDecrease(toDecrease)
	}
}

func (ft *FileTransferImpl) watchRTT(nbMillisec int) {
	sleepDuration := time.Duration(nbMillisec) * time.Millisecond
	ticker := time.NewTicker(sleepDuration)

	go func() {
		for {
			select {
			case <-ticker.C:
				ft.updateRTT()
			case <-ft.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (ft *FileTransferImpl) AskChunk(conn ShosetConn, message *msg.FileMessage, sendRate bool) {
	rate := 0
	message.Rate = rate
	ft.SendMessage(*message, conn)
}

func (ft *FileTransferImpl) ReceiveChunk(conn ShosetConn, message *msg.FileMessage) error {
	fileLeecher, ok := ft.FileLeechers.Load(message.FileUUID)
	if !ok {
		return fmt.Errorf("can't find leecher for file %s", message.FileName)
	}
	err := fileLeecher.(*FileLeecher).ReceiveChunk(conn, message.Begin, message.Length, message.ChunkData)
	return err
}

func (ft *FileTransferImpl) ReceiveAskChunk(conn ShosetConn, message *msg.FileMessage) error {
	ft.MissingDataConn.Store(conn, message.MissingLength)
	if ft.IsAuthorised(conn) {
		origin, ok := ft.FileSeeders.Load(message.FileUUID)
		if !ok { // there is no seeder for this file
			origin, ok = ft.FileLeechers.Load(message.FileUUID)
			if !ok {
				syncFile, err := ft.Library.GetFile(message.FileUUID)
				if err != nil {
					return fmt.Errorf("can't find file %s", message.FileUUID)
				}
				// careful : need to be thread safe
				fmt.Println("ReceiveAskChunk", message.FileName, "create new seeder")
				err = ft.addNewSeeder(syncFile, conn, message)
				return err
			}

			leecher := origin.(*FileLeecher)
			if leecher.haveChunk(message.Begin, message.Length) {
				//fmt.Println("I have the chunk", message.Begin)
				err := leecher.sendBlock(conn, message.Begin, message.Length)
				conn.GetLogger().Trace().Msgf("leecher send block %d,%d,%v", message.Begin, message.Length, err)
				return err
			} else {
				return fmt.Errorf("i don't have the chunk %d", message.Begin)
			}
		} else {
			err := origin.(*FileSeeder).sendBlock(conn, message.Begin, message.Length)
			conn.GetLogger().Trace().Msgf("seeder send block %d,%d,%v", message.Begin, message.Length, err)
			return err
		}
	} else {
		return fmt.Errorf("not authorised")
	}
}

func (ft *FileTransferImpl) addNewSeeder(syncFile SyncFile, conn ShosetConn, message *msg.FileMessage) error {
	ft.m.Lock()
	origin, ok := ft.FileSeeders.Load(syncFile.GetUUID())
	var newSeeder *FileSeeder
	if ok {
		newSeeder = origin.(*FileSeeder)
	} else {
		newSeeder = new(FileSeeder)
		newSeeder.InitSeeder(syncFile, ft)
		ft.FileSeeders.Store(syncFile.GetUUID(), newSeeder)
	}
	ft.m.Unlock()
	err := newSeeder.sendBlock(conn, message.Begin, message.Length)
	return err
}

func (ft *FileTransferImpl) ReceiveAskInfoMessage(conn ShosetConn, message *msg.FileMessage) {
	syncFile, err := ft.Library.GetFile(message.FileUUID)
	if err != nil {
		conn.GetLogger().Warn().Msgf("%v", err)
		return
	}
	ft.SendMessage(syncFile.GetFullInfoMsg(), conn)
}

func (ft *FileTransferImpl) GetMissingLength() int64 {
	ft.m.Lock()
	defer ft.m.Unlock()
	return ft.missingLength
}

func (ft *FileTransferImpl) DecreaseMissingLength(length int) {
	ft.m.Lock()
	defer ft.m.Unlock()
	ft.missingLength = Max64(ft.missingLength-int64(length), 0)
}

// keep asking info if we don't have this file
// to launch in a goroutine
func (ft *FileTransferImpl) AskInfoFile(conn ShosetConn, fileState FileState) {
	ft.m.Lock()

	leecher, ok := ft.FileLeechers.Load(fileState.UUID)
	if ok { // if we already have a leacher for this file, we don't need to ask info
		if leecher.(*FileLeecher).GetHash() == fileState.Hash { // if the hash is the same, we don't need to ask info
			leecher.(*FileLeecher).InitDownload(conn)
			ft.m.Unlock()
			return
		}
	}

	ft.GetInfoMap[fileState.UUID] = fileState.Hash

	// we ask info for this file
	ft.m.Unlock()
	count := 0
	for count < 2 { //  we ask 2 times, after we give up
		ft.m.Lock()
		hash, ok := ft.GetInfoMap[fileState.UUID]
		ft.m.Unlock()
		if !ok || hash != fileState.Hash { // if the hash is not the same, we don't need to ask info
			return
		}
		// we need to keep asking info
		message := msg.FileMessage{MessageName: "askInfo",
			FileName: fileState.Name,
			FileHash: fileState.Hash,
			FileUUID: fileState.UUID,
		}
		message.InitMessageBase()
		ft.SendMessage(message, conn)
		time.Sleep(10 * time.Second)
		count++
	}
}

func (ft *FileTransferImpl) checkRequestTimeout() {
	ft.FileLeechers.Range(func(_, value interface{}) bool {
		value.(*FileLeecher).CheckRequestTimeout()
		return true
	})
}

func (ft *FileTransferImpl) watchRequestsTimeout(nbMilliSec int) {
	sleepDuration := time.Duration(nbMilliSec) * time.Millisecond
	ticker := time.NewTicker(sleepDuration)

	go func() {
		for {
			select {
			case <-ticker.C:
				ft.checkRequestTimeout()
			case <-ft.stop:
				ticker.Stop()
				return

			}
		}
	}()
}

func (ft *FileTransferImpl) InitLeecher(syncFile SyncFile, conn ShosetConn) {
	leecher, ok := ft.FileLeechers.Load(syncFile.GetUUID())
	if !ok { // we don't have a leecher for this file
		newFileLeecher := NewFileLeecher(syncFile, ft)
		ft.AddFileLeecher(newFileLeecher)

	} else {
		// we modify the leecher (for instance when there is another version and we didn't finished the download)
		leecher.(*FileLeecher).UpdateLeeching()
	}
	syncFile.SetStatus("downloading")
}

func (ft *FileTransferImpl) ReceiveMessage(message *msg.FileMessage, conn ShosetConn) {
	var messageQueue *MessageQueue
	messageQueueI, ok := ft.ReceiveQueue.Load(conn)
	if !ok {
		// we do the same but inside a lock to avoid multiple go routine to create the same message queue
		ft.m.Lock()
		messageQueueI, ok := ft.ReceiveQueue.Load(conn)
		if !ok {
			messageQueue = NewMessageQueue()
			go ft.HandleReceiveMessage(messageQueue, conn)
			ft.ReceiveQueue.Store(conn, messageQueue)
		} else {
			messageQueue = messageQueueI.(*MessageQueue)
		}
		ft.m.Unlock()
	} else {
		messageQueue = messageQueueI.(*MessageQueue)
	}
	messageQueue.Push(message)
}

func (ft *FileTransferImpl) HandleReceiveMessage(messageQueue *MessageQueue, conn ShosetConn) {

	for {
		select {
		case <-messageQueue.GetChan():
			message := messageQueue.Pop()
			if message == nil {
				return
			}
			err := ft.HandleReceiveMessageFromQueue(*message, conn)
			if err != nil {
				conn.GetLogger().Warn().Msgf("%v", err)
				// TODO : change the way to handle errors
			}

		case <-ft.stop:
			return
		}
	}
}

func (ft *FileTransferImpl) HandleReceiveMessageFromQueue(fileMessage msg.FileMessage, c ShosetConn) error {
	switch fileMessage.MessageName {
	case "authorised":
		ft.ReceiveAuthorisedMessage(c)
	case "unauthorised":
		ft.ReceiveUnauthorisedMessage(c)
	case "congestion":
		ft.ReceiveCongestionMessage(c)
	case "askInfo":
		ft.ReceiveAskInfoMessage(c, &fileMessage)
	case "interested":
		ft.ReceiveInterestedMessage(c, &fileMessage)
	case "notInterested":
		ft.ReceiveInterestedMessage(c, &fileMessage)
	case "have":
		ft.ReceiveHaveMessage(c, &fileMessage)
	case "sendBitfield":
		ft.ReceiveBitfieldMessage(c, &fileMessage)
	case "askBitfield":
		ft.ReceiveAskBitfieldMessage(c, &fileMessage)

	case "sendInfo":
		fileState := FileState{
			UUID:          fileMessage.FileUUID,
			Name:          fileMessage.FileName,
			Hash:          fileMessage.FileHash,
			HashMap:       fileMessage.FileHashMap,
			Version:       fileMessage.FileVersion,
			LastOperation: ToOperation(fileMessage.FileOperation),
		}
		// we receive some information about a file (probably new or have been modified)
		ft.m.Lock()
		syncFile, err := ft.Library.GetFile(fileMessage.FileUUID)
		if err != nil { // we don't have this file in our library
			if len(fileMessage.FileHashMap) == 0 {
				// we ask for more info
				ft.m.Unlock()
				go ft.AskInfoFile(c, fileState)
			} else {
				// we create the Copy file to be ready to receive the chunks
				CopyFile, err := NewEmptyFile(ft.Library.GetDir(), RealToCopyPath(fileMessage.FilePath), fileMessage.FileName, fileMessage.FileSize, fileMessage.FileHash, fileMessage.FileVersion, fileMessage.FileHashMap)
				if err != nil {
					return err
				}
				// create the syncFile and add it to the library
				syncFile, err := ft.Library.CreateFile(CopyFile, fileMessage.FileUUID)
				if err != nil {
					return err
				}
				ft.m.Unlock()
				// we start the download

				ft.InitLeecher(syncFile, c)

				// normally a new leecher is created
				leecher, ok := ft.FileLeechers.Load(fileMessage.FileUUID)
				if !ok {
					return fmt.Errorf("leecher not found for file %s", fileMessage.FileUUID)
				}
				leecher.(*FileLeecher).InitDownload(c)
				shortInfo := syncFile.GetShortInfoMsg()
				ft.Broadcast(&shortInfo)
			}
		} else { // we already have this file in our library
			ft.m.Unlock()
			syncFile.UpdateFile(fileState, c)
		}

	case "sendChunk":
		// we receive a chunk of the file
		// for now : no verification if we asked it (potential security issue)
		err := ft.ReceiveChunk(c, &fileMessage)
		if err != nil {
			ft.Logger.Error().Err(err).Msg("error while receiving chunk")
		}

	case "askChunk":
		// we receive a request for a chunk of the file
		// we send it back
		_, ok := ft.AuthorisedConn.Load(c)
		if !ok {
			ft.AddUploadConn(c)
		}
		err := ft.ReceiveAskChunk(c, &fileMessage)
		if err != nil {
			ft.Logger.Error().Err(err).Msg("error while receiving askChunk")
			return err
		}
	case "sendLibrary":
		// someone notify us its library
		// we should check if it's from the same Lname

		//c.Logger.Debug().Msg("sendLibrary")
		listFileState := make([]FileState, 0)
		for _, fileStateMsg := range fileMessage.Library {
			listFileState = append(listFileState, ToFileState(fileStateMsg))
		}
		// we update our library
		ft.Library.UpdateLibrary(listFileState, c)
	case "askLibraryLocked":
		// someone ask us if we have a locked the library
		answer := ft.ExternalCommands.IsLocked(&fileMessage)
		answer.FileHash = ft.Library.GetHash()
		ft.SendMessage(answer, c)
	case "answerLibraryLocked":
		// someone answer us if we have a locked the library
		//fmt.Println("comparing hash library", fileMessage.FileHash, ft.Library.GetHash(), fileMessage.AnswerLocked)
		if fileMessage.FileHash != ft.Library.GetHash() {
			// we have a different library
			c.GetLogger().Info().Msg("Received Librarylock message but we have a different library")
			fileMessage.AnswerLocked = true
			libMsg, err := ft.Library.GetMessageLibrary()
			if err != nil {
				return err
			}
			ft.SendMessage(*libMsg, c)
		}
		ft.ExternalCommands.ReceiveAnswer(&fileMessage, c.GetRemoteAddress(), ft.nbConn)
	default:
		return errors.New("Unknown message name for file : " + fileMessage.MessageName)
	}
	return nil
}

func (ft *FileTransferImpl) SendMessage(message msg.FileMessage, conn ShosetConn) {
	if message.MessageName == "sendChunk" {
		message.MissingLength = ft.missingLength
	}

	var messageQueue *MessageQueue
	messageQueueI, ok := ft.SendQueue.Load(conn)
	if !ok {
		// we do the same but inside a lock to avoid multiple go routine to create the same message queue
		ft.m.Lock()
		messageQueueI, ok := ft.SendQueue.Load(conn)
		if !ok {
			//fmt.Println(conn.GetLocalAddress(), "creating message queue", conn.GetRemoteAddress())
			messageQueue = NewMessageQueue()
			go ft.HandleSendMessage(messageQueue, conn)
			ft.SendQueue.Store(conn, messageQueue)
		} else {
			messageQueue = messageQueueI.(*MessageQueue)
		}
		ft.m.Unlock()
	} else {
		messageQueue = messageQueueI.(*MessageQueue)
	}
	messageQueue.Push(&message)
}

func (ft *FileTransferImpl) HandleSendMessage(messageQueue *MessageQueue, conn ShosetConn) {
	for {
		select {
		case <-messageQueue.GetChan():
			message := messageQueue.Pop()
			if message == nil {
				return
			}
			err := ft.HandleSendMessageFromQueue(*message, conn)
			if err != nil {
				fmt.Println(err)
				// TODO : change the way to handle errors
			}

		case <-ft.stop:
			return
		}
	}
}

func (ft *FileTransferImpl) HandleSendMessageFromQueue(fileMessage msg.FileMessage, c ShosetConn) error {
	return c.SendMessage(fileMessage)
}

func (ft *FileTransferImpl) SetExternalCommands(ec *ExternalCommands) {
	ft.ExternalCommands = ec
}

func (ft *FileTransferImpl) SetNbConn(nbConn int) {
	ft.m.Lock()
	defer ft.m.Unlock()
	ft.nbConn = nbConn
}

func (ft *FileTransferImpl) GetLibrary() FileLibrary {
	return ft.Library
}

func (ft *FileTransferImpl) Broadcast(message *msg.FileMessage) {
	ft.broadcast(message)
}

func (ft *FileTransferImpl) UserPush(message *msg.FileMessage) {
	ft.userMessageQueue.Push(message, "", "")
}

func (ft *FileTransferImpl) DeleteLeecher(fileUUID string) {
	ft.FileLeechers.Delete(fileUUID)
}

func (ft *FileTransferImpl) GetLeecher(fileUUID string) *FileLeecher {
	leecher, ok := ft.FileLeechers.Load(fileUUID)
	if !ok {
		return nil
	}
	return leecher.(*FileLeecher)
}

func (ft *FileTransferImpl) WriteRecords() {
	ft.LogRecords.writeRecords()
}

func (ft *FileTransferImpl) GetLogger() *zerolog.Logger {
	return &ft.Logger
}

func (ft *FileTransferImpl) GetReceiveQueue(conn *ShosetConn) *MessageQueue {
	messageQueueI, ok := ft.ReceiveQueue.Load(*conn)
	if !ok {
		return nil
	}
	return messageQueueI.(*MessageQueue)
}
