package server

import (
	"net"
	"log"
	"bufio"
	"sync"
	"github.com/lietu/godistlock/messages"
)

type Relay struct {
	Server        *Server
	RelayId       string
	Connection    net.Conn
	outgoing      chan *OutMsg
	alive         bool
	responseQueue map[string]chan messages.Message
	closeMutex    *sync.Mutex
	responseMutex *sync.Mutex
}

func (r *Relay) Close() {
	// This could end up getting called because of various reasons
	r.closeMutex.Lock()
	defer r.closeMutex.Unlock()

	if r.alive {
		r.alive = false
		r.Connection.Close()
		close(r.outgoing)

		//for lock := range r.heldLocks {
		//	r.Server.LockManager.Release(r.ClientId, lock)
		//}
		//r.heldLocks = map[string]bool{}
	}
}

func (r *Relay) Error(message string) {
	msg, _ := messages.NewClientErrResponse(message)
	r.Outgoing(msg.ToBytes())
	r.Close()
}

func (r *Relay) Outgoing(data []byte) {
	om := OutMsg{
		data,
		make(chan bool),
	}

	r.outgoing <- &om
	<-om.Done
}

func (r *Relay) HandleOutgoing() {
	// Read until channel is closed
	for outgoing := range r.outgoing {
		log.Printf("%s <- %s", r.RelayId, outgoing.Data)
		r.Connection.Write(outgoing.Data)
		r.Connection.Write([]byte("\n"))
		outgoing.Done <- true
	}
	log.Printf("%s outgoing queue closed", r.RelayId)
}

func (r *Relay) HandleHello(msg *messages.RelayIncomingHello) {
	log.Printf("%s HELLO from version %s", r.RelayId, msg.Version)

	// TODO: Register to relay manager

	out, err := messages.NewRelayHowdy([]string{msg.Nonce, r.Server.Id, r.Server.Version})

	if err != nil {
		log.Fatalln(err)
	}

	r.Outgoing(out.ToBytes())
}

func (r *Relay) clearNonce(nonce string) {
	responseQueue := map[string]chan messages.Message{}
	for n, receiver := range r.responseQueue {
		if n != nonce {
			responseQueue[n] = receiver
		}
	}

	r.responseQueue = responseQueue
}

func (r *Relay) gotResponse(msg messages.RelayResponse) {
	r.responseMutex.Lock()
	defer r.responseMutex.Unlock()

	nonce := msg.GetNonce()

	if receiver, ok := r.responseQueue[nonce]; ok {
		r.clearNonce(nonce)
		go func() {
			receiver <- msg.(messages.Message)
		}()
	}
}

func (r *Relay) Incoming(src []byte) {
	keyword, msg, err := messages.LoadMessage("relay", src)

	if err != nil {
		r.Error(err.Error())
		return
	}

	if messages.IsRelayResponse(keyword) {
		r.gotResponse(msg.(messages.RelayResponse))
		return
	}

	switch msg := msg.(type) {
	case *messages.RelayIncomingHello:
		r.HandleHello(msg)
	default:
		r.Error("Invalid keyword")
		r.Close()
	}
}

func (r *Relay) Expect(nonce string, onReceive func(messages.Message)) {
	if DEBUG {
		log.Printf("Relay %s waiting for message with nonce %s", r.RelayId, nonce)
	}

	r.responseMutex.Lock()
	defer r.responseMutex.Unlock()

	receiver := make(chan messages.Message)

	go func() {
		onReceive(<-receiver)
	}()

	r.responseQueue[nonce] = receiver
}

func (r *Relay) DoHello(onComplete func()) {
	nonce := r.Server.GetNonceString()
	r.Expect(nonce, func(msg messages.Message) {
		hello := msg.(*messages.RelayHowdy)

		r.RelayId = hello.Id

		onComplete()
	})

	msg, err := messages.NewRelayIncomingHello([]string{r.Server.Id, r.Server.Version, nonce})

	if err != nil {
		log.Fatalln(err)
	}

	r.Outgoing(msg.ToBytes())
}

func (r *Relay) Run() {
	log.Printf("Processing relay connection %s", r.RelayId)

	go r.HandleOutgoing()

	scanner := bufio.NewScanner(r.Connection)
	for scanner.Scan() {
		// If this data is ever used for longer than a single iteration it
		// must be copied.
		line := scanner.Bytes()
		log.Printf("%s -> %s", r.RelayId, string(line[:]))
		r.Incoming(line)
	}

	// Nothing more to read from the connection, so I guess it was closed
	r.Close()

	log.Printf("%s connection closed", r.RelayId)
}

func NewRelay(server *Server, connection net.Conn) *Relay {
	r := Relay{}

	r.Server = server
	r.Connection = connection
	r.closeMutex = &sync.Mutex{}
	r.outgoing = make(chan *OutMsg)
	r.responseMutex = &sync.Mutex{}
	r.responseQueue = map[string]chan messages.Message{}

	if connection != nil {
		r.RelayId = connection.RemoteAddr().String()
	} else {
		r.RelayId = NewUUID()
	}

	return &r
}
