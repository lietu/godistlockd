package messages

// `HELLO <nonce> <id> <version>` -> Hi, I'm <id> running <version>
// `GIVE <nonce> <fence>` -> Here you go, you now have the lock
// `LOCK <nonce> <fence>` -> Yes, lock <lock> is locked, this is the <fence> token
// `NO <nonce>` -> Lock <lock> is not locked
// `STATS <nonce> <name> <value>` -> Stats response
// `STATSEND <nonce>` -> All stats responses have been sent
// `ERR <msg>` -> System error, you will be disconnected, maybe try another server

type ClientHelloResponse struct {
	Nonce   string
	Id      string
	Version string
}

type ClientOutgoingGive struct {
	Nonce string
	Fence string
}

type ClientErrResponse struct {
	Message string
}

func (msg *ClientHelloResponse) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		msg.Id,
		msg.Version,
	}

	return ToBytes("HELLO", args)
}

func (msg *ClientOutgoingGive) ToBytes() []byte {
	args := []string{
		msg.Nonce,
		msg.Fence,
	}

	return ToBytes("GIVE", args)
}

func (msg *ClientErrResponse) ToBytes() []byte {
	args := []string{
		msg.Message,
	}

	return ToBytes("ERR", args)
}

func NewClientOutgoingHello(nonce string, id string, version string) Message {
	m := ClientHelloResponse{}
	m.Nonce = nonce
	m.Id = id
	m.Version = version

	return &m
}

func NewClientOutgoingGive(nonce string, fence string) Message {
	m := ClientOutgoingGive{}
	m.Nonce = nonce
	m.Fence = fence

	return &m
}

func NewClientErrResponse(reason string) (msg Message, err error) {
	err = nil

	m := ClientErrResponse{}
	m.Message = reason

	msg = &m

	return
}
