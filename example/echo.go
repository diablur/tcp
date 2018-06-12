package main

import (
	"log"
	"time"

	"github.com/diablur/tcp-go"
)

func main() {
	srv := tcp.NewTCPServer("localhost:9001", &EchoCallback{}, &EchoProtocol{})
	srv.SetReadTimeOut(time.Second * 10)
	log.Println("start listen...")
	log.Println(srv.ListenAndServe())
	//select {}
}
