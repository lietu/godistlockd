package server

import (
	"github.com/lietu/godistlock/messages"
	"time"
	"sync"
	"net"
	"log"
)

type RelayConnections map[string]*Relay
type RelayList []*Relay
type MessageList []messages.Message

var WAIT_TIMEOUT = time.Second

type RelayManager struct {
	Server           *Server
	quitChan         chan bool
	relayAddresses   []string
	relayConnections RelayConnections
	serverIds        map[string]string
	serverMutex      *sync.Mutex
}

func (rm *RelayManager) Stop() {
	rm.quitChan <- true
}

func (rm *RelayManager) GetRelayConnections() (relays RelayList) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	relays = RelayList{}
	for _, r := range rm.relayConnections {
		relays = append(relays, r)
	}

	return
}

func waitForMessage(nonce string, relay *Relay, out chan messages.Message) {
	lock := sync.Mutex{}
	sent := false

	relay.Expect(nonce, func(result messages.Message) {
		lock.Lock()
		defer lock.Unlock()

		if !sent {
			sent = true
			out <- result
		}
	})

	go func() {
		time.Sleep(WAIT_TIMEOUT)

		lock.Lock()
		defer lock.Unlock()
		if !sent {
			sent = true
			out <- nil
		}
	}()
}

func (rm *RelayManager) GetRelayResponses(request messages.RelayMessage) (results MessageList) {
	relays := rm.GetRelayConnections()
	count := len(relays)

	responses := make(chan messages.Message)

	for _, relay := range relays {
		nonce := rm.Server.GetNonceString()
		waitForMessage(nonce, relay, responses)
		request.SetNonce(nonce)
		go relay.SendBytes(request.ToBytes())
	}

	results = MessageList{}
	for i := 0; i < count; i++ {
		results = append(results, <-responses)
	}

	return
}

func (rm *RelayManager) getMissingRelays() []string {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	missing := []string{}

	// Addresses that I don't know the server ID for
	for _, addr := range rm.relayAddresses {
		if _, ok := rm.serverIds[addr]; !ok {
			missing = append(missing, addr)
		}
	}

	// + Addresses for servers I know ID for, but are not connected
	for addr, id := range rm.serverIds {
		if _, ok := rm.relayConnections[id]; !ok {
			if id != rm.Server.Id {
				missing = append(missing, addr)
			}
		}
	}

	return missing
}

func (rm *RelayManager) setServerId(addr string, serverId string) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	rm.serverIds[addr] = serverId
}

func (rm *RelayManager) setRelay(relay *Relay) bool {
	if relay.RelayId == rm.Server.Id {
		// Connection to self
		return false
	}

	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	if _, ok := rm.relayConnections[relay.RelayId]; ok {
		if rm.relayConnections[relay.RelayId].Alive {
			return false
		}
	}

	rm.relayConnections[relay.RelayId] = relay

	return true
}

func (rm *RelayManager) connect(addr string) {
	log.Printf("Connecting to relay %s", addr)

	// Initiate connection to target address
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)

		// Failures happen, try again later
		return
	}

	log.Printf("Outgoing relay connection established with %s", addr)

	// Start up new relay handler
	r := NewRelay(rm.Server, conn)
	go r.Run()

	// TODO: Track open connections per destination address as well

	// Perform HELLO <-> HELLO exchange
	r.DoHello(func () {
		log.Printf("Finished saying hellos with %s (%s)", addr, r.RelayId)
		// Update server address<->ID map
		rm.setServerId(addr, r.RelayId)

		if !rm.setRelay(r) {
			log.Printf("Already had a connection with %s, disconnecting", r.RelayId)
			r.Close()
		}
	})
}

func (rm *RelayManager) checkRelays() {
	missing := rm.getMissingRelays()

	for _, addr := range missing {
		go rm.connect(addr)
	}
}

func (rm *RelayManager) Run() {
	checkRelaysInterval := time.Millisecond * 5000
	rm.checkRelays()

	for {
		select {
		case <-rm.quitChan:
			return
		case <-time.After(checkRelaysInterval):
			rm.checkRelays()

			if !rm.Server.Testing {
				break
			}

			go func() {
				msg, _ := messages.NewRelayIncomingHello([]string{rm.Server.Id, rm.Server.Version, "nonce"})

				start := time.Now()
				responses := rm.GetRelayResponses(msg.(messages.RelayMessage))
				duration := time.Since(start)

				log.Printf("Got %d responses in %f s", len(responses), float32(duration) / float32(time.Second))
				for _, response := range responses {
					log.Printf("%+v", response)
				}
			}()
		}
	}
}

func NewRelayManager(server *Server) *RelayManager {
	rm := RelayManager{}
	rm.Server = server
	rm.quitChan = make(chan bool)
	rm.relayAddresses = server.GetRelayAddresses()
	rm.serverMutex = &sync.Mutex{}
	rm.serverIds = map[string]string{}
	rm.relayConnections = RelayConnections{}

	return &rm
}

