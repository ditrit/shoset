package file

import (
	"fmt"
	"testing"
	"time"

	"os"

	"github.com/ditrit/shoset"
)

var newFileContent string = "Another content for test "

func createFile(fileName string) *File{
	file := File{}
	file.m.Lock()
	file.Status = "Empty"

	file.Name = fileName
	//file.Path = filepath.Dir(path)

	//fmt.Println("(NewFile) file.Path : ",file.Path)
	// var err error
	// file.Data, err = os.ReadFile(path)

	file.Data= []byte(newFileContent+" : "+fileName)

	file.Status = "ready"
	file.m.Unlock()
	return &file
}

func prepareContext(t *testing.T){
	os.Chdir("..") //Change the current durectory

	home,err := os.UserHomeDir()
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	err = os.RemoveAll(home+"/.shoset") // Remove .shoset folder before running
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	err = os.RemoveAll("./test_files/destination/") // Delete and recreate the folder destination for transfer tests
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}
	
	err = os.Mkdir("./test_files/destination",0777)
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}
}

func TestPrepareContext(t *testing.T) {
	prepareContext(t)
}

func TestWaitFile(t *testing.T) {
	fmt.Println(os.Getwd())
	prepareContext(t)
	fmt.Println(os.Getwd())	
	
	cl1 := shoset.NewShoset("cl", "cl") //Maitresse du système
	cl1.InitPKI("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("localhost:8002", "localhost:8001", "join") // we join it to our first socket

	time.Sleep(1 * time.Second) //Wait for connexion

	file1 := createFile("test1.txt")

	fmt.Println("File 1 : ",file1)
	//file2 := createFile("test2.txt")

	//Fonction spécifique qui envoi des paquet fait à la main ??

	transfer_tx1 := file1.NewFileTransferTx(cl2, "127.0.0.1:8001")
	fmt.Println("Data to be transfered :", file1)
	fmt.Println("transfer1 : ", transfer_tx1.String())
	go transfer_tx1.HandleTransfer() //Start the transfer

	// receive data :
	transfer_rx1 := NewFileTransferRx(cl1, "127.0.0.1:8002")
	received1 := transfer_rx1.WaitFile()

	fmt.Println("file1.Data",file1.Data)
	fmt.Println("received1.Data",received1.Data)

	fmt.Println("string(received1.Data)",string(received1.Data))

	if !(string(received1.Data) == string(file1.Data)) {
		t.Errorf("Wrong content for file 1")
	}

	/*
		-Importer 2 fichiers
		-Changer les contenues
		-Lancer les 2 transferts
		-Vérifier les contenues
	*/

	//t.Errorf("Test not implemented yet.")

}
