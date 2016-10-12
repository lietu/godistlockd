package messages

import "time"

// `HELLO <version> <nonce>` -> Hi, I'm a client running version <version>
// `ON <lock> <timeout> <nonce>` -> Wait until you get lock, keep locked until timeout, will return a token for fencing
// `OFF <lock> <nonce>` -> Release lock
// `TRY <lock> <timeout> <nonce>` -> Check if you can get lock, get it if you can, will return a token for fencing
// `REFRESH <lock> <fence> <timeout> <nonce>` -> I want to keep this lock for a bit longer
// `IS <lock> <nonce>` -> Check if the lock is engaged, returns fence token (nonce) if it is
// `STATS <nonce>` -> Get count of locks and other stats about the system

type ClientIncomingHello struct {
	Version string
	Nonce   string
}

type ClientIncomingOn struct {
	Lock    string
	Timeout time.Duration
	Nonce   string
}

type ClientIncomingOff struct {
	Lock    string
	Nonce   string
}

// ClientHelloMessage

func (msg *ClientIncomingHello) ToBytes() []byte {
	args := []string{
		msg.Version,
		msg.Nonce,
	}

	return ToBytes("HELLO", args)
}

// ClientOnMessage

func (msg *ClientIncomingOn) ToBytes() []byte {
	args := []string{
		msg.Lock,
		DurationToString(msg.Timeout),
		msg.Nonce,
	}

	return ToBytes("ON", args)
}

// ClientIncomingOff

func (msg *ClientIncomingOff) ToBytes() []byte {
	args := []string{
		msg.Lock,
		msg.Nonce,
	}

	return ToBytes("OFF", args)
}

// Constructors

func NewClientIncomingHello(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ErrInvalidMessage
		return
	}

	m := ClientIncomingHello{}
	m.Version = args[0]
	m.Nonce = args[1]

	msg = &m

	return
}

func NewClientIncomingOn(args []string) (msg Message, err error) {
	if len(args) != 3 {
		err = ErrInvalidMessage
		return
	}

	m := ClientIncomingOn{}
	m.Lock = args[0]
	m.Timeout, err = StringToDuration(args[1])

	if err != nil {
		return
	}

	m.Nonce = args[2]

	msg = &m

	return
}

func NewClientIncomingOff(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ErrInvalidMessage
		return
	}

	m := ClientIncomingOff{}
	m.Lock = args[0]
	m.Nonce = args[1]

	msg = &m

	return
}

func init() {
	RegisterMessageType("client_incoming", "HELLO", NewClientIncomingHello)
	RegisterMessageType("client_incoming", "ON", NewClientIncomingOn)
	RegisterMessageType("client_incoming", "OFF", NewClientIncomingOff)
}
