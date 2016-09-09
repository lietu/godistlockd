package messages

// `HELLO <id> <version> <nonce>` -> I'm server <id> running <version>
// `PROP <lock> <nonce>` -> I propose locking, please give me your lock status
// `SCHED <lock> <nonce>` -> We have quorum, nobody is locked, prep to lock
// `COMM <lock> <timeout> <nonce>` -> Commit lock with X timeout
// `OFF <lock> <nonce>` -> Release lock if it was held by the source relay


type RelayIncomingHello struct {
	Id      string
	Version string
	Nonce   string
}

// RelayIncomingHello

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
		err = ERR_INVALID_MESSAGE
		return
	}

	m := RelayIncomingHello{}
	m.Id = args[0]
	m.Version = args[1]
	m.Nonce = args[2]

	msg = &m

	return
}

func init() {
	RegisterMessageType("relay", "HELLO", NewRelayIncomingHello)
}
