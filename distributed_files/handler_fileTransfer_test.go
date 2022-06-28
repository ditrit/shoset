package file

import (
	"fmt"
	"path"
	"testing"
	"time"

	"os"
	"sync"

	//"github.com/rs/zerolog"
	//"github.com/rs/zerolog/log"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

var newFileContent string = "Another content for test "

func createFile(fileName string) *File {
	file := File{}
	file.m.Lock()
	file.Status = "Empty"

	file.Name = fileName
	//file.Path = filepath.Dir(path)

	//fmt.Println("(NewFile) file.Path : ",file.Path)
	// var err error
	// file.Data, err = os.ReadFile(path)

	file.Data = []byte(newFileContent + " : " + fileName)

	file.Status = "ready"
	file.m.Unlock()
	return &file
}

func prepareContext(t *testing.T) {
	//Change the current durectory to acces some folders
	workingDir, err := os.Getwd()
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}
	if path.Base(workingDir) == "distributed_files" {
		os.Chdir("..")
	}

	// Deleting some data generated by last tests
	home, err := os.UserHomeDir()
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	err = os.RemoveAll(home + "/.shoset") // Remove .shoset folder before running
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	err = os.RemoveAll("./test_files/destination/") // Delete and recreate the folder destination for transfer tests
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	err = os.Mkdir("./test_files/destination", 0777)
	if err != nil {
		t.Errorf(err.Error())
		fmt.Println(err)
	}

	t.Log("Done with seting curent directory to ./shoset and deleting results of earlier tests.")
}

func TestPrepareContext(t *testing.T) {
	prepareContext(t)
}

func TestWaitFile(t *testing.T) {
	//fmt.Println(os.Getwd())
	prepareContext(t)
	//fmt.Println(os.Getwd())

	//zerolog.

	var wg sync.WaitGroup

	cl1 := shoset.NewShoset("cl", "cl") //Maitresse du système
	cl1.InitPKI("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("localhost:8002", "localhost:8001", "join") // we join it to our first socket

	time.Sleep(1 * time.Second) //Wait for connexion

	testfiles_tx := []*File{}
	testfiles_rx := []*File{}

	for i := 0; i < 2; i++ {
		testfiles_tx = append(testfiles_tx, createFile("file"+fmt.Sprint(i)))
	}

	t.Log("testfiles_tx", testfiles_tx)

	iterator := msg.NewIterator(cl1.Queue["fileChunk"])

	for _, file := range testfiles_tx {
		//Sender :
		wg.Add(1)
		go func() {
			transfer_tx := file.NewFileTransferTx(cl2, "127.0.0.1:8001")
			transfer_tx.HandleTransfer() //Start the transfer
			defer wg.Done()
		}()

		time.Sleep(10 * time.Millisecond)

		//Receiver :
		//wg.Add(1)
		//Avoir un itérateur commun
		//Imposer le nom du fichier à récupérer : fonctionne
		//Consommation des messages

		// Pas de reception simultanée

		//go func() {
		transfer_rx := NewFileTransferRx(cl1, "127.0.0.1:8002")
		received := transfer_rx.WaitFile(iterator) //iterator
		//received := transfer_rx.WaitFileName(iterator,"file"+fmt.Sprint(i))
		testfiles_rx = append(testfiles_rx, received)
		//defer wg.Done()
		//}()

		//time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()

	for i, file := range testfiles_rx {
		t.Log(file.String())

		fmt.Println("FileName : ", file.Name, "Data : ", file.Data)

		if !(string(testfiles_tx[i].Data) == string(testfiles_rx[i].Data)) {
			t.Errorf("Wrong content for file" + fmt.Sprint(i))
		}
	}
}
