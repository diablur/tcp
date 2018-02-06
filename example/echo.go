package main

import (
	"log"
	"time"

	"tcp"
)

func main() {
	srv := tcp.NewTCPServer("localhost:9001", &EchoCallback{}, &EchoProtocol{})
	srv.SetReadTimeOut(time.Second * 10)
	log.Println("start listen...")
	log.Println(srv.ListenAndServe())
	//select {}
}
