package messages

import (
	"time"
)

//
// `HELLO <id> <version> <nonce>` -> I'm server <id> running <version>
// 

type RelayIncomingHello struct {
	Id      string
	Version string
	Nonce   string
}

func (msg *RelayIncomingHello) ToBytes() []byte {
	args := []string{
		msg.Id,
		msg.Version,
		msg.Nonce,
	}

	return ToBytes("HELLO", args)
}

func (msg *RelayIncomingHello) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayIncomingHello) GetNonce() string {
	return msg.Nonce
}

func NewRelayIncomingHello(args []string) (msg Message, err error) {
	if len(args) != 3 {
		err = ErrInvalidMessage
		return
	}

	m := RelayIncomingHello{}
	m.Id = args[0]
	m.Version = args[1]
	m.Nonce = args[2]

	msg = &m

	return
}


//
// `PROP <lock> <nonce>` -> I propose locking, please give me your lock status
// 

type RelayIncomingProp struct {
	Lock  string
	Nonce string
}

func (msg *RelayIncomingProp) ToBytes() []byte {
	args := []string{
		msg.Lock,
		msg.Nonce,
	}

	return ToBytes("PROP", args)
}

func (msg *RelayIncomingProp) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayIncomingProp) GetNonce() string {
	return msg.Nonce
}

func NewRelayIncomingProp(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ErrInvalidMessage
		return
	}

	m := RelayIncomingProp{}
	m.Lock = args[0]
	m.Nonce = args[1]

	msg = &m

	return
}


//
// `SCHED <lock> <nonce>` -> We have quorum, nobody is locked, prep to lock
// 

type RelayIncomingSched struct {
	Lock  string
	Nonce string
}

func (msg *RelayIncomingSched) ToBytes() []byte {
	args := []string{
		msg.Lock,
		msg.Nonce,
	}

	return ToBytes("SCHED", args)
}

func (msg *RelayIncomingSched) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayIncomingSched) GetNonce() string {
	return msg.Nonce
}

func NewRelayIncomingSched(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ErrInvalidMessage
		return
	}

	m := RelayIncomingSched{}
	m.Lock = args[0]
	m.Nonce = args[1]

	msg = &m

	return
}


//
// `COMM <lock> <timeout> <nonce>` -> Commit lock with X timeout
// 

type RelayIncomingComm struct {
	Lock    string
	Timeout time.Duration
	Nonce   string
}

func (msg *RelayIncomingComm) ToBytes() []byte {
	args := []string{
		msg.Lock,
		DurationToString(msg.Timeout),
		msg.Nonce,
	}

	return ToBytes("COMM", args)
}

func (msg *RelayIncomingComm) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayIncomingComm) GetNonce() string {
	return msg.Nonce
}

func NewRelayIncomingComm(args []string) (msg Message, err error) {
	if len(args) != 3 {
		err = ErrInvalidMessage
		return
	}

	m := RelayIncomingComm{}
	m.Lock = args[0]
	m.Timeout, err = StringToDuration(args[1])
	m.Nonce = args[2]

	if err != nil {
		err = ErrInvalidMessage
		return
	}

	msg = &m

	return
}


//
// `OFF <lock> <nonce>` -> Release lock if it was held by the source relay
//

type RelayIncomingOff struct {
	Lock    string
	Nonce   string
}

func (msg *RelayIncomingOff) ToBytes() []byte {
	args := []string{
		msg.Lock,
		msg.Nonce,
	}

	return ToBytes("OFF", args)
}

func (msg *RelayIncomingOff) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayIncomingOff) GetNonce() string {
	return msg.Nonce
}

func NewRelayIncomingOff(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ErrInvalidMessage
		return
	}

	m := RelayIncomingOff{}
	m.Lock = args[0]
	m.Nonce = args[1]

	msg = &m

	return
}


// -----

func init() {
	RegisterMessageType("relay", "HELLO", NewRelayIncomingHello)
	RegisterMessageType("relay", "PROP", NewRelayIncomingProp)
	RegisterMessageType("relay", "SCHED", NewRelayIncomingSched)
	RegisterMessageType("relay", "COMM", NewRelayIncomingComm)
	RegisterMessageType("relay", "OFF", NewRelayIncomingOff)
}
