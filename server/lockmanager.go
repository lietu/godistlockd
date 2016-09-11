package server

import (
	"github.com/aristanetworks/goarista/monotime"
	"time"
	"log"
	"strconv"
)

var DEBUG = true

const (
	TYPE_GET = iota
	TYPE_TRY
	TYPE_CHECK
	TYPE_RELEASE
)

type LockQueue map[string][]*LockRequest
type Locks map[string]*Lock

type Lock struct {
	Fence    string
	Expires  uint64
	ClientId string
}

type LockRequest struct {
	Name     string
	ClientId string
	Timeout  time.Duration
	Type     int
	Done     chan *Lock
}

type LockManager struct {
	requestChan chan *LockRequest
	quitChan    chan bool
	locks       Locks
}

func (lm *LockManager) Stop() {
	lm.quitChan <- true
}

func (lm *LockManager) GetLock(clientId string, name string, timeout time.Duration) *Lock {
	receiver := NewLockReceiver()
	receiver.ClientId = clientId
	receiver.Name = name
	receiver.Timeout = timeout
	receiver.Type = TYPE_GET

	lm.requestChan <- receiver

	return <-receiver.Done
}

func (lm *LockManager) TryGet(clientId string, name string, timeout time.Duration) *Lock {
	receiver := NewLockReceiver()
	receiver.ClientId = clientId
	receiver.Name = name
	receiver.Timeout = timeout
	receiver.Type = TYPE_TRY

	lm.requestChan <- receiver

	return <-receiver.Done
}

func (lm *LockManager) IsLocked(name string) string {
	receiver := NewLockReceiver()
	receiver.Name = name
	receiver.Type = TYPE_CHECK

	lm.requestChan <- receiver

	result := <-receiver.Done

	if result == nil {
		return ""
	} else {
		return result.Fence
	}
}

func (lm *LockManager) Release(clientId string, name string) {
	receiver := NewLockReceiver()
	receiver.Name = name
	receiver.ClientId = clientId
	receiver.Type = TYPE_RELEASE

	lm.requestChan <- receiver
	<-receiver.Done
}

func (lm *LockManager) giveLock(receiver *LockRequest) {
	lock := Lock{}

	// Use monotonic clocks, time.Now() can jump around
	lock.Expires = monotime.Now() + uint64(receiver.Timeout)
	lock.Fence = NewFence()
	lock.ClientId = receiver.ClientId

	lm.locks[receiver.Name] = &lock

	if DEBUG {
		log.Printf("Giving lock %s away until %d", receiver.Name, lock.Expires)
	}

	receiver.Done <- &lock
}

func (lm *LockManager) isLocked(name string) (clientId string) {
	clientId = ""
	lock, ok := lm.locks[name]
	if ok && lock.Expires > monotime.Now() {
		clientId = lock.ClientId
	}
	return
}

func (lm *LockManager) release(clientId string, name string) {
	_, ok := lm.locks[name]

	if ok {
		newLocks := Locks{}

		for n, lock := range lm.locks {
			if n == name && lock.ClientId == clientId {
				if DEBUG {
					log.Printf("Lock %s was released.", n)
				}
				continue
			}

			newLocks[n] = lock
		}

		lm.locks = newLocks
	}
}

func appendToQueue(queue *LockQueue, receiver *LockRequest) {
	if _, ok := (*queue)[receiver.Name]; !ok {
		(*queue)[receiver.Name] = []*LockRequest{}
	}
	(*queue)[receiver.Name] = append((*queue)[receiver.Name], receiver)
}

func (lm *LockManager) handleGet(clientId string, request *LockRequest) (result bool) {
	result = false
	if clientId == "" {
		if DEBUG {
			log.Printf("Lock %s was free, so giving it as requested.", request.Name)
		}

		lm.giveLock(request)
		result = true
	} else if clientId == request.ClientId {
		if DEBUG {
			log.Printf("Client asked to re-establish lock %s", request.Name)
		}

		request.Done <- lm.locks[request.Name]
		result = true
	}

	return
}

func (lm *LockManager) handleTry(clientId string, request *LockRequest) {
	if clientId == "" {
		if DEBUG {
			log.Printf("Lock %s was free, so giving it as requested.", request.Name)
		}

		lm.giveLock(request)
	} else if clientId == request.ClientId {
		if DEBUG {
			log.Printf("Client asked to re-establish lock %s", request.Name)
		}

		request.Done <- lm.locks[request.Name]
	} else {
		if DEBUG {
			log.Printf("Lock %s was taken, and request did not want to wait for it.", request.Name)
		}
		request.Done <- nil
	}
}

func (lm *LockManager) checkQueue(queue LockQueue) LockQueue {
	newQueue := LockQueue{}
	for _, requests := range queue {
		for _, request := range requests {
			locked := lm.isLocked(request.Name)

			if locked == "" {
				if DEBUG {
					log.Printf("Lock %s has expired, giving it to the next one in queue.", request.Name)
				}
				lm.giveLock(request)
			} else {
				appendToQueue(&newQueue, request)
			}
		}
	}

	return newQueue
}

func (lm *LockManager) Run() {
	queueCheckInterval := time.Millisecond * 10
	queue := LockQueue{}

	for {
		select {
		case request := <-lm.requestChan:
			clientId := lm.isLocked(request.Name)

			if request.Type == TYPE_GET {
				if !lm.handleGet(clientId, request) {
					if DEBUG {
						log.Printf("Lock %s was taken, and request wants to wait for it.", request.Name)
					}
					appendToQueue(&queue, request)
				}
			} else if request.Type == TYPE_TRY {
				lm.handleTry(clientId, request)
			} else if request.Type == TYPE_CHECK {
				if clientId == "" {
					request.Done <- nil
				} else {
					request.Done <- lm.locks[request.Name]
				}
			} else if request.Type == TYPE_RELEASE {
				if clientId != "" {
					lm.release(request.ClientId, request.Name)
					request.Done <- nil
				}
			}

		case <-time.After(queueCheckInterval):
			queue = lm.checkQueue(queue)

		case <-lm.quitChan:
			if DEBUG {
				log.Println("LockManager quitting")
			}
			return
		}
	}
}

func NewFence() string {
	// TODO: This should figure out the correct fence to use in relay network
	return strconv.FormatUint(monotime.Now(), 10)
}

func NewLockReceiver() *LockRequest {
	lr := LockRequest{}
	lr.Done = make(chan *Lock)
	return &lr
}

func NewLockManager() *LockManager {
	lm := LockManager{}

	lm.locks = map[string]*Lock{}

	lm.requestChan = make(chan *LockRequest)
	lm.quitChan = make(chan bool)

	go lm.Run()

	return &lm
}
