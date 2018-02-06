package tcp

import (
	"io"
)

type Protocol interface {
	ReadPacket(reader io.Reader) (Packet, error)
	WritePacket(writer io.Writer, msg Packet) error
}
