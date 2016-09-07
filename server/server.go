package server

import (
	"net"
	"log"
	"fmt"
	"strconv"
)

type LockStatus map[string]Lock;

type Server struct {
	Id           string
	Version      string
	nonceChan    chan uint64
	lockStatus   LockStatus
	LockManager  *LockManager
	RelayManager *RelayManager
}

func (s *Server) GetNonce() uint64 {
	return <-s.nonceChan
}

func (s *Server) GetNonceString() string {
	return strconv.FormatUint(s.GetNonce(), 10)
}

func (s *Server) nonceGenerator() {
	var i uint64
	for i = 1; ; i += 1 {
		s.nonceChan <- i
	}
}

func (s *Server) GetRelayAddresses() []string {
	return []string{"localhost:20000"}
	return []string{"localhost:20000", "localhost:20001", "localhost:20002"}
}

func NewServer() *Server {
	s := Server{}
	s.nonceChan = make(chan uint64)
	s.lockStatus = LockStatus{}
	s.LockManager = NewLockManager()
	// TODO: Ensure settings are loaded before this line
	s.RelayManager = NewRelayManager(&s)

	return &s
}

func startClient(server *Server, connection net.Conn) {
	c := NewClient(server, connection)
	c.Run()
}

func startRelay(server *Server, connection net.Conn) {
	r := NewRelay(server, connection)
	r.Run()
}

func (s *Server) clientListener(port int) {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	log.Printf("Started listening for client connections on TCP port %d", port)

	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go startClient(s, conn)
	}
}

func (s *Server) relayListener(port int) {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	log.Printf("Started listening for relay connections on TCP port %d", port)

	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go startRelay(s, conn)
	}
}

func (s *Server) Run(clientPort int, relayPort int) {
	go s.nonceGenerator()
	go s.LockManager.Run()
	go s.RelayManager.Run()
	go s.clientListener(clientPort)
	s.relayListener(relayPort)
}
