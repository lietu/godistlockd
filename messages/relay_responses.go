package messages

import (
	"strconv"
)

var responseTypes = []string{
	"HOWDY",
	"STAT",
	"ACK",
	"CONF",
}

//
// `HOWDY <nonce> <id> <version>` -> Hi, I'm <id> running <version>//
//

type RelayHowdy struct {
	Nonce   string
	Id      string
	Version string
}

func (msg *RelayHowdy) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		msg.Id,
		msg.Version,
	}

	return ToBytes("HOWDY", args)
}

func (msg *RelayHowdy) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayHowdy) GetNonce() string {
	return msg.Nonce
}

func NewRelayHowdy(args []string) (msg Message, err error) {
	if len(args) != 3 {
		err = ERR_INVALID_MESSAGE
		return
	}

	m := RelayHowdy{}
	m.Nonce = args[0]
	m.Id = args[1]
	m.Version = args[2]

	msg = &m

	return
}

//
// `STAT <nonce> <status>` -> Response to PROP: status 0 = ok, 1 = held by this server, 2 = held by another relay
//

type RelayStat struct {
	Nonce   string
	Status  int
}

func (msg *RelayStat) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		strconv.Itoa(msg.Status),
	}

	return ToBytes("STAT", args)
}

func (msg *RelayStat) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayStat) GetNonce() string {
	return msg.Nonce
}

func NewRelayStat(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ERR_INVALID_MESSAGE
		return
	}

	m := RelayStat{}
	m.Nonce = args[0]
	m.Status, err = strconv.Atoi(args[1])

	if err != nil {
		err = ERR_INVALID_MESSAGE
		return
	}

	msg = &m

	return
}

//
// `ACK <nonce> <status>` -> Acknowledging SCHED: status 1 = ok, 0 = err
//

type RelayAck struct {
	Nonce   string
	Status  int
}

func (msg *RelayAck) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		strconv.Itoa(msg.Status),
	}

	return ToBytes("ACK", args)
}

func (msg *RelayAck) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayAck) GetNonce() string {
	return msg.Nonce
}

func NewRelayAck(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ERR_INVALID_MESSAGE
		return
	}

	m := RelayAck{}
	m.Nonce = args[0]
	m.Status, err = strconv.Atoi(args[1])

	if err != nil {
		err = ERR_INVALID_MESSAGE
		return
	}

	msg = &m

	return
}

//
// `CONF <nonce> <status>` -> Confirming commit 1/0 = ok/err
//

type RelayConf struct {
	Nonce   string
	Status  int
}

func (msg *RelayConf) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		strconv.Itoa(msg.Status),
	}

	return ToBytes("CONF", args)
}

func (msg *RelayConf) SetNonce(nonce string) {
	msg.Nonce = nonce
}

func (msg *RelayConf) GetNonce() string {
	return msg.Nonce
}

func NewRelayConf(args []string) (msg Message, err error) {
	if len(args) != 2 {
		err = ERR_INVALID_MESSAGE
		return
	}

	m := RelayConf{}
	m.Nonce = args[0]
	m.Status, err = strconv.Atoi(args[1])

	if err != nil {
		err = ERR_INVALID_MESSAGE
		return
	}

	msg = &m

	return
}

// -----

func IsRelayResponse(t string) bool {
	for _, msgType := range responseTypes {
		if t == msgType {
			return true
		}
	}

	return false
}

func init() {
	RegisterMessageType("relay", "HOWDY", NewRelayHowdy)
	RegisterMessageType("relay", "STAT", NewRelayStat)
	RegisterMessageType("relay", "ACK", NewRelayAck)
	RegisterMessageType("relay", "CONF", NewRelayConf)
}
