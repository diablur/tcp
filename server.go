package tcp

import (
	"errors"
	"log"
	"net"
	"time"
)

const (
	readChanSize  = 100
	writeChanSize = 100
)

var removeChan = make(chan string, 100)

type TCPServer struct {
	addr         string
	listener     *net.TCPListener
	callback     CallBack
	protocol     Protocol
	exitChan     chan struct{}
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	bucket       *ConnBucket
}

func NewTCPServer(addr string, callback CallBack, protocol Protocol) *TCPServer {
	return &TCPServer{
		addr:     addr,
		callback: callback,
		protocol: protocol,
		bucket:   NewConnBucket(),
		exitChan: make(chan struct{}),
	}
}

func (srv *TCPServer) ListenAndServe() error {
	addr, err := net.ResolveTCPAddr("tcp4", srv.addr)
	if err != nil {
		return err
	}
	srv.listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}
	err = srv.Serve(srv.listener)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TCPServer) Serve(l *net.TCPListener) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Serve error", r)
		}
		_ = srv.listener.Close()
	}()

	go srv.bucket.removeClosedTCPConnLoop()

	var tempDelay time.Duration
	for {
		select {
		case <-srv.exitChan:
			return errors.New("Server Closed")
		default:
		}
		conn, err := srv.listener.AcceptTCP()
		if err != nil {
			if n, ok := err.(net.Error); ok && n.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			log.Println("listen error:", err.Error())
			return err
		}
		tempDelay = 0
		tcpConn, _ := srv.newTCPConn(conn, srv.callback, srv.protocol)
		tcpConn.setReadTimeOut(srv.readTimeOut)
		tcpConn.setWriteTimeOut(srv.writeTimeOut)
		srv.bucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (srv *TCPServer) newTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) (*TCPConnection, error) {
	c := NewTCPConnection(conn, callback, protocol)
	e := c.Serve()
	return c, e
}

func (srv *TCPServer) Connect(ip string, callback CallBack, protocol Protocol) (*TCPConnection, error) {
	addr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	return srv.newTCPConn(conn, callback, protocol)
}

func (srv *TCPServer) Close() {
	defer func() {
		err := srv.listener.Close()
		if err != nil {
			log.Printf("err close listener: %s\n", err.Error())
		}
	}()
	for _, c := range srv.bucket.GetAll() {
		if !c.IsClosed() {
			c.Close(0)
		}
	}
}

func (srv *TCPServer) GetAllTCPConn() []*TCPConnection {
	var conns []*TCPConnection
	for _, conn := range srv.bucket.GetAll() {
		conns = append(conns, conn)
	}
	return conns
}

func (srv *TCPServer) GetTCPConn(key string) *TCPConnection {
	return srv.bucket.Get(key)
}

func (srv *TCPServer) SetReadTimeOut(t time.Duration) {
	srv.readTimeOut = t
}

func (srv *TCPServer) SetWriteTimeOut(t time.Duration) {
	srv.writeTimeOut = t
}
