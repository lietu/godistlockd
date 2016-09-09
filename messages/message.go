package messages

type Message interface {
	ToBytes() []byte
}

type RelayMessage interface {
	SetNonce(nonce string)
	GetNonce() string
	ToBytes() []byte
}
