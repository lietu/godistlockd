package server

import (
	"github.com/lietu/godistlock/messages"
	"time"
	"sync"
	"net"
	"log"
	"math"
)

type RelayConnections map[string]*Relay
type RelayList []*Relay
type MessageList []messages.Message

var WAIT_TIMEOUT = time.Second

type RelayManager struct {
	Server                *Server
	quitChan              chan bool
	relayAddresses        []string
	relayConnections      RelayConnections
	pendingConnections    []string
	serverIds             map[string]string
	serverMutex           *sync.Mutex
	quorumNeed            int
	canHaveQuorum         bool
	relayAddressesHasSelf bool
}

func (rm *RelayManager) Stop() {
	rm.quitChan <- true
}

func (rm *RelayManager) GetRelayConnections() (relays RelayList) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	relays = RelayList{}
	for _, r := range rm.relayConnections {
		if r.Alive {
			relays = append(relays, r)
		}
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
		nonce := relay.Nonce.String()
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

	// Filter out already pending connections
	notPending := []string{}
	for _, addr := range missing {
		add := true
		for _, a := range rm.pendingConnections {
			if a == addr {
				add = false
				break
			}
		}

		if add {
			notPending = append(notPending, addr)
		}
	}

	return notPending
}

func (rm *RelayManager) setServerId(addr string, serverId string) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	rm.serverIds[addr] = serverId
}

func (rm *RelayManager) updateQuorum() {
	connections := len(rm.GetRelayConnections())
	rm.canHaveQuorum = (connections >= rm.quorumNeed)

	rm.Server.RelayManagerReady(rm.canHaveQuorum)
}

func (rm *RelayManager) SetRelay(relay *Relay) bool {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	if relay.RelayId == rm.Server.Id {
		if !rm.relayAddressesHasSelf {
			rm.relayAddressesHasSelf = true
			rm.quorumNeed = calculateQuorum(len(rm.relayAddresses))
			go rm.updateQuorum()
		}
		// Connection to self
		return false
	}

	if _, ok := rm.relayConnections[relay.RelayId]; ok {
		if rm.relayConnections[relay.RelayId].Alive {
			return false
		}
	}

	rm.relayConnections[relay.RelayId] = relay
	go rm.updateQuorum()

	return true
}

func (rm *RelayManager) addPendingConnection(addr string) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	rm.pendingConnections = append(rm.pendingConnections, addr)
}

func (rm *RelayManager) removePendingConnection(addr string) {
	rm.serverMutex.Lock()
	defer rm.serverMutex.Unlock()

	pendingConnections := []string{}
	for _, a := range rm.pendingConnections {
		if a == addr {
			continue
		}

		pendingConnections = append(pendingConnections, a)
	}

	rm.pendingConnections = pendingConnections
}

func (rm *RelayManager) connect(addr string) {
	log.Printf("Connecting to relay %s", addr)
	rm.addPendingConnection(addr)

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

	// Perform HELLO <-> HELLO exchange
	r.DoHello(func() {
		log.Printf("Finished saying hellos with %s (%s)", addr, r.RelayId)
		// Update server address<->ID map
		rm.setServerId(addr, r.RelayId)

		if !rm.SetRelay(r) {
			log.Printf("Already had a connection with %s, disconnecting", r.RelayId)
			r.Close()
		}

		rm.removePendingConnection(addr)
	})
}

func (rm *RelayManager) checkRelays() {
	missing := rm.getMissingRelays()

	for _, addr := range missing {
		go rm.connect(addr)
	}
}

func (rm *RelayManager) Run() {
	checkRelaysInterval := time.Millisecond
	rm.checkRelays()

	for {
		select {
		case <-rm.quitChan:
			return
		case <-time.After(checkRelaysInterval):
			rm.checkRelays()

			log.Printf("%d relays connected", len(rm.GetRelayConnections()))

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

func calculateQuorum(servers int) int {
	return (int)(math.Ceil((float64)(servers) / 100.0 * 50.01))
}

func NewRelayManager(server *Server) *RelayManager {
	rm := RelayManager{}
	rm.Server = server
	rm.quitChan = make(chan bool)
	rm.relayAddresses = server.GetRelayAddresses()
	rm.serverMutex = &sync.Mutex{}
	rm.serverIds = map[string]string{}
	rm.relayConnections = RelayConnections{}
	rm.pendingConnections = []string{}
	rm.quorumNeed = calculateQuorum(len(rm.relayAddresses) + 1)
	rm.relayAddressesHasSelf = false
	rm.canHaveQuorum = false

	return &rm
}

