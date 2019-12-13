package main

import (
	"crypto/tls"
	"fmt"
	"net"
)

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// GSServer : first test on event client socket
type GSServer struct {
	config    *tls.Config
	address   string
	gSClients map[string]*tls.Conn
}

// NewGSServer : constructor
func NewGSServer(address string) *GSServer {
	s := new(GSServer)
	cert, _ := tls.LoadX509KeyPair(certPath, keyPath)
	s.config = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	s.address = address

	return s
}

// Run : handler for the socket
func (s *GSServer) Run() error {
	fmt.Print("GSServer.Run launched\n")

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		fmt.Print("GSServer initialized\n")
		return err
	}
	defer listener.Close()

	for {
		connUnenc, err := listener.Accept()
		if err != nil {
			fmt.Printf("server: accept %s", err)
			break
		}
		conn := tls.Server(connUnenc, s.config)
		s.gSClients[conn.RemoteAddr().String()] = conn
		fmt.Printf("GSServer : accepted from %s", conn.RemoteAddr())
		go s.handleConnection(conn)

	}
	return nil
}

// HandleConnection : handler for the socket
func (s *GSServer) handleConnection(conn *tls.Conn) {
	fmt.Printf("GSServer.HandleConn launched for %s\n", conn.RemoteAddr())
	remAddr := conn.RemoteAddr().String()
	buffer := make([]byte, 1024)
	for {
		fmt.Print("Waiting data...")
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Failed to read:", err.Error())
			delete(s.gSClients, remAddr)
			break
		}
		fmt.Printf("Received and echoing %s", string(buffer[:bytesRead]))
		_, err = conn.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Println("Failed to write:", err.Error())
			delete(s.gSClients, remAddr)
			break
		}
		fmt.Printf("successed in writing !")
	}
}

func main() {
	test := NewGSServer("localhost:8100")
	test.Run()
}
