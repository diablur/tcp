package tcp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var ()

type TCPConnection struct {
	callback     CallBack
	protocol     Protocol
	conn         *net.TCPConn
	readChan     chan Packet
	writeChan    chan Packet
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	exitChan     chan bool
	closeOnce    sync.Once
	exitFlag     bool
	wg           sync.WaitGroup
}

func NewTCPConnection(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConnection {
	c := &TCPConnection{
		conn:      conn,
		callback:  callback,
		protocol:  protocol,
		readChan:  make(chan Packet, readChanSize),
		writeChan: make(chan Packet, writeChanSize),
		exitChan:  make(chan bool),
		exitFlag:  false,
		wg:        sync.WaitGroup{},
	}
	return c
}

func (c *TCPConnection) Serve() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("tcp conn(%s) Serve error, %v\n ", c.GetRemoteIPAddress(), r)
		}
	}()
	if c.callback == nil || c.protocol == nil {
		err := fmt.Errorf("callback and protocol are not allowed to be nil")
		c.Close(0)
		return err
	}
	c.callback.OnConnected(c)
	c.wg.Add(3)
	go c.readLoop()
	go c.writeLoop()
	go c.handleLoop()
	return nil
}

func (c *TCPConnection) readLoop() {
	defer func() {
		_ = recover()
		c.wg.Done()
		c.Close(0)
	}()
	for {
		select {
		case <-c.exitChan:
			return
		default:
			if c.readTimeOut > 0 {
				_ = c.conn.SetReadDeadline(time.Now().Add(c.readTimeOut))
			}
			p, err := c.protocol.ReadPacket(c.conn)
			if err != nil {
				if err != io.EOF {
					c.callback.OnError(err)
				}
				return
			}
			c.readChan <- p
		}
	}
}

func (c *TCPConnection) writeLoop() {
	defer func() {
		_ = recover()
		c.wg.Done()
		c.Close(0)
	}()
	for {
		select {
		case <-c.exitChan:
			return
		case pkt := <-c.writeChan:
			if c.writeTimeOut > 0 {
				_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeOut))
			}
			if err := c.protocol.WritePacket(c.conn, pkt); err != nil {
				c.callback.OnError(err)
				return
			}
		}
	}
}

func (c *TCPConnection) handleLoop() {
	defer func() {
		_ = recover()
		c.wg.Done()
		c.Close(0)
	}()
	for {
		select {
		case <-c.exitChan:
			return
		case p := <-c.readChan:
			c.callback.OnMessage(c, p)
		}
	}
}

func (c *TCPConnection) ReadPacket() (Packet, error) {
	if c.protocol == nil {
		return nil, errors.New("no protocol implements")
	}
	return c.protocol.ReadPacket(c.conn)
}

func (c *TCPConnection) WritePacket(p Packet) error {
	if c.IsClosed() {
		return errors.New("use of closed connection")
	}
	select {
	case c.writeChan <- p:
		return nil
	default:
		return errors.New("send buffer is full")
	}
}

func (c *TCPConnection) Close(flag int) {
	c.closeOnce.Do(func() {
		close(c.exitChan)
		c.wg.Wait()
		close(c.writeChan)
		close(c.readChan)
		c.exitFlag = true
		_ = c.conn.Close()
		removeChan <- c.GetRemoteAddr().String()
		if c.callback != nil && flag == 0 {
			c.callback.OnDisconnected(c)
		}
	})
}

func (c *TCPConnection) GetRawConn() *net.TCPConn {
	return c.conn
}

func (c *TCPConnection) IsClosed() bool {
	return c.exitFlag
}

func (c *TCPConnection) GetLocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TCPConnection) GetLocalIPAddress() string {
	return strings.Split(c.GetLocalAddr().String(), ":")[0]
}

func (c *TCPConnection) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConnection) GetRemoteIPAddress() string {
	return strings.Split(c.GetRemoteAddr().String(), ":")[0]
}

func (c *TCPConnection) setReadTimeOut(t time.Duration) {
	c.readTimeOut = t
}

func (c *TCPConnection) setWriteTimeOut(t time.Duration) {
	c.writeTimeOut = t
}
