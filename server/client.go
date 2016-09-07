package server

import (
	"log"
	"net"
	"bufio"
	"sync"
	"github.com/lietu/godistlock/messages"
)

type Client struct {
	Server     *Server
	ClientId   string
	Connection net.Conn
	alive      bool
	outgoing   chan *OutMsg
	closeMutex *sync.Mutex
	heldLocks  map[string]bool
}

type OutMsg struct {
	Data []byte
	Done chan bool
}

func (c *Client) addLock(name string) {
	c.heldLocks[name] = true
}

func (c *Client) removeLock(name string) {
	heldLocks := map[string]bool{}

	for n := range c.heldLocks {
		if n == name {
			continue
		}

		heldLocks[name] = true
	}

	c.heldLocks = heldLocks
}

/// For unit tests
func (c *Client) GetHeldLocks() map[string]bool {
	return c.heldLocks
}

func (c *Client) Close() {
	// This could end up getting called because of various reasons
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if c.alive {
		c.alive = false
		c.Connection.Close()
		close(c.outgoing)

		for lock := range c.heldLocks {
			c.Server.LockManager.Release(c.ClientId, lock)
		}
		c.heldLocks = map[string]bool{}
	}
}

func (c *Client) Error(message string) {
	msg, _ := messages.NewClientErrResponse(message)
	c.Outgoing(msg.ToBytes())
	c.Close()
}

func (c *Client) Outgoing(data []byte) {
	log.Printf("%s -> %s", c.ClientId, string(data))

	om := OutMsg{
		data,
		make(chan bool),
	}

	c.outgoing <- &om
	<-om.Done
}

func (c *Client) HandleHello(msg *messages.ClientIncomingHello) {
	log.Printf("%s HELLO from version %s", c.ClientId, msg.Version)

	// TODO: Check client version is supported

	out := messages.NewClientOutgoingHello(msg.Nonce, c.Server.Id, c.Server.Version)
	c.Outgoing(out.ToBytes())
}

func (c *Client) HandleOn(msg *messages.ClientIncomingOn) {
	log.Printf("%s requesting lock %s", c.ClientId, msg.Lock)

	lock := c.Server.LockManager.GetLock(c.ClientId, msg.Lock, msg.Timeout)
	out := messages.NewClientOutgoingGive(msg.Nonce, lock.Fence)
	c.Outgoing(out.ToBytes())

	c.addLock(msg.Lock)
}

func (c *Client) HandleOff(msg *messages.ClientIncomingOff) {
	log.Printf("%s releasing lock %s", c.ClientId, msg.Lock)

	c.Server.LockManager.Release(c.ClientId, msg.Lock)

	c.removeLock(msg.Lock)
}

func (c *Client) HandleOutgoing() {
	// Read until channel is closed
	for outgoing := range c.outgoing {
		c.Connection.Write(outgoing.Data)
		c.Connection.Write([]byte("\n"))
		outgoing.Done <- true
	}
	log.Printf("%s outgoing queue closed", c.ClientId)
}

func (c *Client) Incoming(src []byte) {
	_, msg, err := messages.LoadMessage("client_incoming", src)

	if err != nil {
		c.Error(err.Error())
		return
	}

	switch msg := msg.(type) {
	case *messages.ClientIncomingHello:
		c.HandleHello(msg)
	case *messages.ClientIncomingOn:
		c.HandleOn(msg)
	case *messages.ClientIncomingOff:
		c.HandleOff(msg)
	default:
		c.Error("Invalid keyword")
		c.Close()
	}
}

func (c *Client) Run() {
	log.Printf("%s new connection", c.ClientId)

	go c.HandleOutgoing()

	scanner := bufio.NewScanner(c.Connection)
	for scanner.Scan() {
		// If this data is ever used for longer than a single iteration it
		// must be copied.
		line := scanner.Bytes()
		c.Incoming(line)
	}

	// Nothing more to read from the connection, so I guess it was closed
	c.Close()

	log.Printf("%s connection closed", c.ClientId)
}

func NewClient(server *Server, connection net.Conn) *Client {
	c := Client{}
	c.Server = server
	c.Connection = connection
	c.alive = true
	c.outgoing = make(chan *OutMsg)
	c.closeMutex = &sync.Mutex{}
	c.heldLocks = map[string]bool{}

	if connection != nil {
		c.ClientId = connection.RemoteAddr().String()
	} else {
		c.ClientId = NewUUID()
	}

	return &c
}