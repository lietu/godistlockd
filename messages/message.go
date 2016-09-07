package messages

type Message interface {
	ToBytes() []byte
}
