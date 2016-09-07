package messages

// `HOWDY <nonce> <id> <version>` -> Hi, I'm <id> running <version>
// `STAT <nonce> <status>` -> Response to PROP: status 0 = ok, 1 = held by this server, 2 = held by another relay
// `ACK <nonce> <ok>` -> Acknowledging SCHED: ok 1 = ok, 0 = err
// `CONF <nonce>` -> Confirming commit 1/0 = ok/err

var responseTypes = []string{
	"HOWDY",
	"STAT",
	"ACK",
	"CONF",
}

type RelayResponse interface {
	GetNonce() string
	ToBytes() []byte
}

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
}
