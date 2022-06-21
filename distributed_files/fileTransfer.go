package files

import (
	"fmt"
	"math"

	"github.com/ditrit/shoset"
)

var chunkSize int = 10

type FileTransfer struct {
	sender         *shoset.Shoset
	transferType   string               // "tx" or "rx" (Lock the data of the file for the duration of transfer)
	file           *File                //File to be transfered
	receivedChunks []int                //List of the ids of chunks received
	sources        []*shoset.ShosetConn //List of connexions involved in the transfer
	/*
		List of chunks requested by a connexion
		Requested chunks must also be in received or the file is complete
	*/
	requestedChunks map[*shoset.ShosetConn][]int
}

//destination : adrress (IP:port) of the destination
func (file *File) NewFileTransfer(sender *shoset.Shoset, transferType string, destinationAdress string) FileTransfer {
	var transfer FileTransfer
	transfer.sender = sender
	transfer.transferType = transferType //"tx" or "rx"
	transfer.file = file
	transfer.receivedChunks = []int{}
	transfer.sources = []*shoset.ShosetConn{}
	transfer.requestedChunks = make(map[*shoset.ShosetConn][]int)

	switch transfer.transferType {
	case "tx": //Sending file
		//Finding the adress in the established cons of the sender
		var conn *shoset.ShosetConn

		for _, i := range sender.GetConnsByTypeArray("cl") {
			fmt.Println("i.GetRemoteAddress()", i.GetRemoteAddress())
			if i.GetRemoteAddress() == destinationAdress {
				conn = i
			}
		}

		//Compute the number of chunks required create the sequence containing every chunls and adds them to the list of requested.
		var chunkRange []int
		for i := 0; i < int(math.Ceil((float64(len(transfer.file.Data)) / float64(chunkSize)))); i++ {
			chunkRange = append(chunkRange, i)
		}

		//fmt.Println("len(transfer.file.Data) (NewFileTransfer) : ",len(transfer.file.Data))
		//fmt.Println("chunkRange (NewFileTransfer) : ",chunkRange)

		transfer.requestedChunks[conn] = chunkRange
	case "rx": //Receiving a file
		//Finding the adress in the established cons of the sender
		//var conn *shoset.ShosetConn // Conn sending file

		//transfer.requestedChunks[conn] = []int{}

	}

	return transfer
}

func (transfer *FileTransfer) String() string {
	result := "\nFileTranfer of " + transfer.file.Name + " :\n"
	result += "sender : " + transfer.sender.String() + "\n"
	result += "transferType : " + transfer.transferType + "\n"
	result += "Amount received : " + fmt.Sprint((len(transfer.receivedChunks))) + "\n"
	result += "Sources (adresses) : "
	for _, i := range transfer.sources {
		result += i.GetRemoteAddress() + ", "
	}
	result += "\n"
	result += "Requested ((adresses) : (amount)) : "
	for conn, chunks := range transfer.requestedChunks {
		result += conn.GetRemoteAddress() + " : " + fmt.Sprint((len(chunks))) + ", "
	}
	result += "\n"
	return result
}
