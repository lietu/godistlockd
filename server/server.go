package server

import (
	"net"
	"log"
	"fmt"
	"sync"
	"time"
)

type LockStatus map[string]Lock;

type Server struct {
	Id                  string
	Version             string
	Testing             bool
	lockStatus          LockStatus
	LockManager         *LockManager
	RelayManager        *RelayManager
	clientPort          int
	statusMutex         sync.Mutex
	listeningForClients bool
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
	s.statusMutex = sync.Mutex{}
	s.listeningForClients = false

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

func (s *Server) RelayManagerReady(ready bool) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if ready && !s.listeningForClients {
		s.listeningForClients = true

		log.Println("RelayManager is ready and we can start listening for clients")
		go s.clientListener(s.clientPort)
	}
}

func (s *Server) GetLock(client *Client, name string, timeout time.Duration) {
	
}

func (s *Server) Run(clientPort int, relayPort int) {
	go s.LockManager.Run()
	go s.RelayManager.Run()
	s.clientPort = clientPort
	s.relayListener(relayPort)
}
