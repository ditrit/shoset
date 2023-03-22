package fileSync

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

/*
It stores in a csv file the flowrate of each connection.
Initialy, it was used to store the upload and download rate.
With the new variable used to monitor the flowrate, it is only storing this variable :  the RTT of each request.
*/

type LogRateTransfer struct {
	rateMap map[int64]map[ShosetConn]int // map[conn]map[timestamp][uploadRate, downloadRate]
	connMap map[ShosetConn]bool
	path    string
	file    *os.File
	writer  *csv.Writer
	m       sync.RWMutex

	titleWritten bool
}

func NewLogRateTransfer(path string) *LogRateTransfer {
	return &LogRateTransfer{
		rateMap: make(map[int64]map[ShosetConn]int),
		connMap: make(map[ShosetConn]bool),
		path:    path,
	}
}

func (logRateTransfer *LogRateTransfer) createMonitorFile() {
	logRateTransfer.m.Lock()
	defer logRateTransfer.m.Unlock()
	file, err := os.Create(logRateTransfer.path)
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	logRateTransfer.file = file
	logRateTransfer.writer = csv.NewWriter(file)
}

// append the flowrates to the file.
// it empty the map that srtores this rates
func (logRateTransfer *LogRateTransfer) writeRecords() error {
	logRateTransfer.m.Lock()
	defer logRateTransfer.m.Unlock()
	fmt.Println("writeRecords", logRateTransfer.connMap)
	var err error
	if !logRateTransfer.titleWritten {
		title := []string{"timestamp"}
		for conn := range logRateTransfer.connMap {
			title = append(title, conn.GetRemoteAddress()+"_uploadRate", conn.GetRemoteAddress()+"_downloadRate")
		}
		err = logRateTransfer.writer.Write(title)
		if err != nil {
			return err
		}
		logRateTransfer.writer.Flush()
		logRateTransfer.titleWritten = true
	}
	for timestamp, connMap := range logRateTransfer.rateMap {
		record := []string{strconv.FormatInt(timestamp, 10)}
		count := 0
		for conn := range logRateTransfer.connMap {
			rate, ok := connMap[conn]
			if ok {
				count += rate
				record = append(record, setNA(rate))
			} else {
				record = append(record, "")
			}
		}
		if count == 0 {
			// there is nothing to write
			continue
		}
		err = logRateTransfer.writer.Write(record)
		if err != nil {
			return err
		}
		logRateTransfer.writer.Flush()
	}
	logRateTransfer.rateMap = make(map[int64]map[ShosetConn]int)
	return nil //logRateTransfer.file.Close()
}

func (logRateTransfer *LogRateTransfer) AddRateConn(timestamp int64, conn ShosetConn, RTT int) {
	logRateTransfer.m.Lock()
	defer logRateTransfer.m.Unlock()
	connMap, ok := logRateTransfer.rateMap[timestamp]
	if !ok {
		connMap = make(map[ShosetConn]int)
		logRateTransfer.rateMap[timestamp] = connMap
	}
	connMap[conn] = RTT
	logRateTransfer.connMap[conn] = true
}

func setNA(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}
