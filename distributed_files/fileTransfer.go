package file

import (
	"fmt"
	"math"

	"github.com/ditrit/shoset"
)

var chunkSize int = 10 //Size in byte of chunks

type FileTransfer struct {
	//(Lock the data of the file for the duration of transfer)
	shosetCom      *shoset.Shoset       //Shoset used to communicate for the transfer
	transferType   string               // "tx" (transmit) : send file, "rx" (receive)
	file           *File                //File to be transfered
	receivedChunks []int                //List of the ids of chunks received
	sources        []*shoset.ShosetConn //List of connexions involved in the transfer
	/*
		List of chunks requested by a connexion
		Requested chunks must also be in received or the file is complete
	*/
	requestedChunks map[*shoset.ShosetConn][]int

	expectedChunks map[*shoset.ShosetConn][]int
}

//Create a new FileTransfer object (Transfer is not started.)
//destinationAdress : adrress (IP:port) of the destination
//transferType : tx (transmit) : send file, rx (receive)
func (file *File) NewFileTransferTx(sender *shoset.Shoset, destinationAdress string) FileTransfer {
	var transfer FileTransfer
	transfer.shosetCom = sender
	transfer.transferType = "tx" //"tx" or "rx"
	transfer.file = file
	// Nécessité de l'initialisation ??
	// transfer.receivedChunks = []int{}
	// transfer.sources = []*shoset.ShosetConn{}
	transfer.requestedChunks = make(map[*shoset.ShosetConn][]int) // For the transmiting side (tx)

	transfer.expectedChunks = make(map[*shoset.ShosetConn][]int) // For the receiving side (rx)

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
	return transfer
}

//Create a new FileTransfer object (Transfer is not started.)
//destinationAdress : adrress (IP:port) of the destination
//transferType : tx (transmit) : send file, rx (receive)
func NewFileTransferRx(receiver *shoset.Shoset, originAdress string) FileTransfer {
	var transfer FileTransfer
	transfer.shosetCom = receiver
	transfer.transferType = "rx" //"tx" or "rx"
	transfer.file = NewEmptyFile()
	// Nécessité de l'initialisation ??
	// transfer.receivedChunks = []int{}
	// transfer.sources = []*shoset.ShosetConn{}
	transfer.requestedChunks = make(map[*shoset.ShosetConn][]int) // For the transmiting side (tx)

	transfer.expectedChunks = make(map[*shoset.ShosetConn][]int) // For the receiving side (rx)

	//Finding the adress in the established cons of the sender
	var conn *shoset.ShosetConn

	for _, i := range receiver.GetConnsByTypeArray("cl") {
		fmt.Println("i.GetRemoteAddress()", i.GetRemoteAddress())
		if i.GetRemoteAddress() == originAdress {
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

	transfer.expectedChunks[conn] = chunkRange
	return transfer
}

func (transfer *FileTransfer) String() string {
	result := "\nFileTranfer of " + transfer.file.Name + " :\n"
	result += "sender : " + transfer.shosetCom.String() + "\n"
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
