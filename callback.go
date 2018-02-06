package tcp

type CallBack interface {
	OnConnected(conn *TCPConnection)
	OnMessage(conn *TCPConnection, p Packet)
	OnDisconnected(conn *TCPConnection)
	OnError(err error)
}
