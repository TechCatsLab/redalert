package server

import (
	"net"
	"sync"
	"bytes"
	"fmt"
)

type Client struct {
	srv         *TcpServer
	Conn        *net.TCPConn
	closeOnce   sync.Once
	closeChan   chan struct{}
	receiveChan chan *Msg
}

func NewClient(conn *net.TCPConn, srv *TcpServer) *Client {
	return &Client{
		srv:         srv,
		Conn:        conn,
		closeChan:   make(chan struct{}),
		receiveChan: make(chan *Msg, 100),
	}
}

func (c *Client) Start() {
	defer func() {
		c.close()
	}()

	go c.readMsg()

	for {
		select {
		case <-c.closeChan:
			return
		case msg := <-c.srv.sender:
			if msg.remote == c.getConnRemoteAddr() {
				_, err := c.Conn.Write(msg.content.Bytes())
				if err != nil {
					return
				}
			}
		default:
		}

		data := make([]byte, 1024)

		_, err := c.Conn.Read(data)
		if err != nil {
			return
		}

		c.receiveChan <- &Msg{*bytes.NewBuffer(data), c.getConnRemoteAddr()}
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		c.Conn.Close()
	})
}

func (c *Client) getConnRemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Client) readMsg() {
	for {
		select {
		case msg := <- c.receiveChan:
			fmt.Println("receive msg:", msg)
		default:
		}
	}
}
