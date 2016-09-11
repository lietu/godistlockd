package server

import (
	"net"
	"log"
	"fmt"
)

type LockStatus map[string]Lock;

type Server struct {
	Id           string
	Version      string
	Testing      bool
	lockStatus   LockStatus
	LockManager  *LockManager
	RelayManager *RelayManager
}

func (s *Server) GetRelayAddresses() []string {
	return []string{"localhost:20000", "localhost:20001", "localhost:20002", "localhost:20003"}
}

func NewServer() *Server {
	s := Server{}
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
	go s.LockManager.Run()
	go s.RelayManager.Run()
	go s.clientListener(clientPort)
	s.relayListener(relayPort)
}
