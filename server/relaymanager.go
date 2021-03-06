package server

import (
	"github.com/lietu/godistlockd/messages"
	"time"
	"sync"
	"net"
	"log"
	"math"
	"fmt"
	"math/rand"
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
	connectMutex          *sync.Mutex
	quorumNeed            int
	CanHaveQuorum         bool
	relayAddressesHasSelf bool
	connecting            bool
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
	myId := RELAY_ID_PREFIX + rm.Server.Id
	for addr, id := range rm.serverIds {
		if id == myId {
			// Skip self
			continue
		}

		r, ok := rm.relayConnections[id]
		if !ok || r.Alive == false {
			missing = append(missing, addr)
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
	rm.CanHaveQuorum = (connections >= rm.quorumNeed)

	rm.Server.RelayManagerReady(rm.CanHaveQuorum)
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

func (rm *RelayManager) RelayDisconnected() {
	go rm.updateQuorum()
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
		rm.removePendingConnection(addr)
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
	if rm.connecting {
		return
	}

	rm.connectMutex.Lock()
	rm.connecting = true
	defer rm.connectMutex.Unlock()

	missing := rm.getMissingRelays()
	count := len(missing)

	if count > 0 {
		wg := sync.WaitGroup{}
		wg.Add(count)
		log.Printf("Missing connection to %d relays", len(missing))

		for _, addr := range missing {
			go func(addr string) {
				rm.connect(addr)
				wg.Done()
			}(addr)
		}

		wg.Wait()
	}

	rm.connecting = false
}

func (rm *RelayManager) ProposeLock(name string) bool {
	if !rm.CanHaveQuorum {
		log.Print("Can't have quorum, not gonna propose locking")
		return false
	}

	log.Printf("Proposing locking of %s", name)
	msg, err := messages.NewRelayIncomingProp([]string{name, "nonce"})

	if err != nil {
		log.Fatal("Failed to create outgoing PROP")
	}

	responses := rm.GetRelayResponses(msg.(messages.RelayMessage))

	ok := 0
	for _, response := range responses {
		if response == nil {
			continue
		}

		r := response.(*messages.RelayStat)
		if r.Status == 0 {
			ok += 1
		}
	}

	return ok >= rm.quorumNeed
}

func (rm *RelayManager) SchedLock(name string) bool {
	if !rm.CanHaveQuorum {
		log.Print("Can't have quorum, not gonna request locking")
		return false
	}

	log.Printf("Requesting lock of %s", name)
	msg, err := messages.NewRelayIncomingSched([]string{name, "nonce"})

	if err != nil {
		log.Fatal("Failed to create outgoing SCHED")
	}

	responses := rm.GetRelayResponses(msg.(messages.RelayMessage))

	ok := 0
	for _, response := range responses {
		if response == nil {
			continue
		}

		r := response.(*messages.RelayAck)
		if r.Status == 0 {
			ok += 1
		}
	}

	return ok >= rm.quorumNeed
}

func (rm *RelayManager) CommLock(name string, timeout time.Duration) bool {
	if !rm.CanHaveQuorum {
		log.Print("Can't have quorum, can't commit lock")
		return false
	}

	log.Printf("Committing lock %s", name)
	msg, err := messages.NewRelayIncomingComm([]string{name, messages.DurationToString(timeout), "nonce"})

	if err != nil {
		log.Fatal("Failed to create outgoing COMM")
	}

	responses := rm.GetRelayResponses(msg.(messages.RelayMessage))

	ok := 0
	for _, response := range responses {
		if response == nil {
			continue
		}

		r := response.(*messages.RelayConf)
		if r.Status == 0 {
			ok += 1
		}
	}

	return ok >= rm.quorumNeed
}

func (rm *RelayManager) Run() {
	rm.checkRelays()

	checks := time.Millisecond * 5

	status := time.Now()
	test := time.Now()
	relayCheck := time.Now()

	for {
		select {
		case <-rm.quitChan:
			return
		case <-time.After(checks):
			if time.Since(relayCheck) > time.Second {
				relayCheck = time.Now()
				rm.checkRelays()
			}

			if time.Since(status) > time.Second * 5 {
				status = time.Now()

				connections := rm.GetRelayConnections()
				log.Printf("%d relays connected", len(connections))
				for _, c := range connections {
					log.Printf(" - %s", c.RelayId)
				}
			}

			if time.Since(test) > time.Millisecond {
				test = time.Now()

				if !rm.Server.Testing {
					break
				}
				go func() {
					lock := fmt.Sprintf("mah-lock-%d", rand.Int31())
					log.Printf("%+v", rm.Server.DoLock("janne", lock, time.Minute))
				}()
			}

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
	rm.connectMutex = &sync.Mutex{}
	rm.serverMutex = &sync.Mutex{}
	rm.serverIds = map[string]string{}
	rm.relayConnections = RelayConnections{}
	rm.pendingConnections = []string{}
	rm.quorumNeed = calculateQuorum(len(rm.relayAddresses) + 1)
	rm.relayAddressesHasSelf = false
	rm.CanHaveQuorum = false
	rm.connecting = false

	return &rm
}

