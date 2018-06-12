package main

import (
	"log"

	"github.com/diablur/tcp-go"
)

type EchoCallback struct{}

func (ec *EchoCallback) OnConnected(conn *tcp.TCPConnection) {
	log.Println("new conn: ", conn.GetRemoteAddr())
}
func (ec *EchoCallback) OnMessage(conn *tcp.TCPConnection, p tcp.Packet) {
	log.Printf("receive: %s", string(p.Bytes()))
	conn.WritePacket(p)
}
func (ec *EchoCallback) OnDisconnected(conn *tcp.TCPConnection) {
	log.Printf("%s disconnected\n", conn.GetRemoteAddr())
}

func (ec *EchoCallback) OnError(err error) {
	log.Println("err msg:", err)
}
