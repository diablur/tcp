package tcp

type Packet interface {
	Bytes() []byte
}
