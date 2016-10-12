package server

import (
	"net"
	"log"
	"bufio"
	"sync"
	"github.com/lietu/godistlock/messages"
	"fmt"
	"strconv"
	"time"
)

const RELAY_ID_PREFIX = "relay:"

type Relay struct {
	Server        *Server
	RelayId       string
	Connection    net.Conn
	outgoing      chan *OutMsg
	Alive         bool
	responseQueue map[string]chan messages.Message
	closeMutex    *sync.Mutex
	responseMutex *sync.Mutex
	Nonce         *NonceGenerator
}

func (r *Relay) Close() {
	// This could end up getting called because of various reasons
	r.closeMutex.Lock()
	defer r.closeMutex.Unlock()

	if r.Alive {
		r.Alive = false
		r.Connection.Close()
		close(r.outgoing)
		r.Server.RelayManager.RelayDisconnected()

		//for lock := range r.heldLocks {
		//	r.Server.LockManager.Release(r.ClientId, lock)
		//}
		//r.heldLocks = map[string]bool{}
	}
}

func (r *Relay) Error(message string) {
	msg, _ := messages.NewClientErrResponse(message)
	r.SendBytes(msg.ToBytes())
	log.Printf("Relay encountered error: %s", message)
	r.Close()
}

func (r *Relay) SendBytes(data []byte) {
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
		//log.Printf("%s <- %s", r.RelayId, string(outgoing.Data[:]))
		r.Connection.Write(outgoing.Data)
		r.Connection.Write([]byte("\n"))
		outgoing.Done <- true
	}
	log.Printf("%s outgoing queue closed", r.RelayId)
}

func (r *Relay) HandleHello(msg *messages.RelayIncomingHello) {
	r.RelayId = RELAY_ID_PREFIX + msg.Id
	// Not handling failures here so other server always gets a valid response
	r.Server.RelayManager.SetRelay(r)

	out, err := messages.NewRelayHowdy([]string{msg.Nonce, r.Server.Id, r.Server.Version})

	if err != nil {
		log.Fatalln(err)
	}

	r.SendBytes(out.ToBytes())
}

func (r *Relay) HandleProp(msg *messages.RelayIncomingProp) {
	// 0 = ok, 1 = held by this server, 2 = held by another relay
	status := 0

	// Try to get a preliminary lock
	lock := r.Server.LockManager.TryGet(r.RelayId, msg.Lock, time.Second)

	if lock == nil {
		clientId := r.Server.LockManager.WhoHas(msg.Lock)
		if isRelayId(clientId) {
			status = 2
		} else {
			status = 1
		}
	}

	out, err := messages.NewRelayStat([]string{msg.Nonce, strconv.Itoa(status)})

	if err != nil {
		log.Fatalln(err)
	}

	r.SendBytes(out.ToBytes())
}

func (r *Relay) HandleSched(msg *messages.RelayIncomingSched) {
	// status 0 = ok, 1 = err
	status := 0

	// Refresh preliminary lock
	lock := r.Server.LockManager.TryGet(r.RelayId, msg.Lock, time.Second)

	if lock == nil {
		status = 1
	}

	out, err := messages.NewRelayAck([]string{msg.Nonce, strconv.Itoa(status)})

	if err != nil {
		log.Fatalln(err)
	}

	r.SendBytes(out.ToBytes())
}

func (r *Relay) HandleComm(msg *messages.RelayIncomingComm) {
	// status 0 = ok, 1 = err
	status := 0

	// Establish a firm lock
	lock := r.Server.LockManager.GetLock(r.RelayId, msg.Lock, msg.Timeout)

	if lock == nil {
		status = 1
	}

	out, err := messages.NewRelayConf([]string{msg.Nonce, strconv.Itoa(status)})

	if err != nil {
		log.Fatalln(err)
	}

	r.SendBytes(out.ToBytes())
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

func (r *Relay) gotResponse(msg messages.RelayMessage) {
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
		r.gotResponse(msg.(messages.RelayMessage))
		return
	}

	switch msg := msg.(type) {
	case *messages.RelayIncomingHello:
		r.HandleHello(msg)
	case *messages.RelayIncomingProp:
		r.HandleProp(msg)
	case *messages.RelayIncomingSched:
		r.HandleSched(msg)
	case *messages.RelayIncomingComm:
		r.HandleComm(msg)
	default:
		r.Error(fmt.Sprintf("Unsupported incoming keyword: %s", keyword))
		r.Close()
	}
}

func (r *Relay) Expect(nonce string, onReceive func(messages.Message)) {
	if DEBUG {
		//log.Printf("Relay %s waiting for message with nonce %s", r.RelayId, nonce)
	}

	r.responseMutex.Lock()
	defer r.responseMutex.Unlock()

	receiver := make(chan messages.Message)

	go func() {
		onReceive(<-receiver)
		//log.Printf("Relay %s received msg for nonce %s", r.RelayId, nonce)
	}()

	r.responseQueue[nonce] = receiver
}

func (r *Relay) DoHello(onComplete func()) {
	nonce := r.Nonce.String()
	r.Expect(nonce, func(msg messages.Message) {
		hello := msg.(*messages.RelayHowdy)

		r.RelayId = RELAY_ID_PREFIX + hello.Id

		onComplete()
	})

	msg, err := messages.NewRelayIncomingHello([]string{r.Server.Id, r.Server.Version, nonce})

	if err != nil {
		log.Fatalln(err)
	}

	r.SendBytes(msg.ToBytes())
}

func (r *Relay) Run() {
	log.Printf("Processing relay connection %s", r.RelayId)

	go r.HandleOutgoing()

	scanner := bufio.NewScanner(r.Connection)
	for scanner.Scan() {
		// If this data is ever used for longer than a single iteration it
		// must be copied.
		line := scanner.Bytes()
		//log.Printf("%s -> %s", r.RelayId, string(line[:]))
		r.Incoming(line)
	}

	// Nothing more to read from the connection, so I guess it was closed
	log.Printf("%s connection closed", r.RelayId)
	r.Close()
}

func NewRelay(server *Server, connection net.Conn) *Relay {
	r := Relay{}

	r.Server = server
	r.Alive = true
	r.Connection = connection
	r.closeMutex = &sync.Mutex{}
	r.outgoing = make(chan *OutMsg)
	r.responseMutex = &sync.Mutex{}
	r.responseQueue = map[string]chan messages.Message{}
	r.Nonce = NewNonceGenerator()

	if connection != nil {
		r.RelayId = connection.RemoteAddr().String()
	} else {
		r.RelayId = NewUUID()
	}

	return &r
}

func isRelayId(id string) bool {
	l := len(RELAY_ID_PREFIX)

	if len(id) < l {
		return false
	}

	return id[:l] == RELAY_ID_PREFIX
}
