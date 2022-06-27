package file

import (
	"fmt"
	"testing"
	"time"

	"github.com/ditrit/shoset"
	
)

func TestWaitFile(t *testing.T) {
	cl1 := shoset.NewShoset("cl", "cl") //Maitresse du système
	cl1.InitPKI("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("localhost:8002", "localhost:8001", "join") // we join it to our first socket

	time.Sleep(1 * time.Second) //Wait for connexion

	files_list_1 := NewFiles()
	files_list_1.AddNewFile("./test_files/source/test1.txt")
	files_list_1.AddNewFile("./test_files/source/test2.txt")
	files_list_1.PrintAllFiles() //Print names of all files

	fmt.Println("files_list_1.FilesMap[test1.txt].Data : ", string(files_list_1.FilesMap["test1.txt"].Data))
	//files_list_1.FilesMap["test1.txt"].Data = []byte("Another content for test1") //Replace content of file in memory.
	//fmt.Println("files_list_1.FilesMap[test1.txt].Data : ", string(files_list_1.FilesMap["test1.txt"].Data))

	//Fonction spécifique qui envoi des paquet fait à la main

	transfer_tx := files_list_1.FilesMap["test1.txt"].NewFileTransferTx(cl2, "127.0.0.1:8001")
	fmt.Println("Data to be transfered :", files_list_1.FilesMap["test1.txt"])
	fmt.Println("transfer1 : ", transfer_tx.String())
	go transfer_tx.HandleTransfer() //Start the transfer

	transfer_rx := NewFileTransferRx(cl1, "127.0.0.1:8002")
	received := transfer_rx.WaitFile()
	fmt.Println("Received File :\n", received)
	err_write := received.WriteToDisk("./test_files/destination")

	if err_write != nil {
		fmt.Println(err_write)
	}

	/*
	-Importer 2 fichiers
	-Changer les contenues
	-Lancer les 2 transferts
	-Vérifier les contenues
	*/

	t.Errorf("Test not implemented yet.")

}
