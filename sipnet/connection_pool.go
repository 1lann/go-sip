package sipnet

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"
)

func (l *Listener) getUDPConnFromPool(address net.Addr) *Conn {
	l.udpPoolMutex.Lock()
	defer l.udpPoolMutex.Unlock()
	conn, found := l.udpPool[address.String()]
	if !found {
		conn = &Conn{
			Transport:        "udp",
			Listener:         l,
			Conn:             l.udpListener,
			Address:          address,
			UdpReceiver:      make(chan []byte),
			Closed:           false,
			Locked:           false,
			WriteBuffer:      new(bytes.Buffer),
			ReadMessage:      make(chan interface{}),
			LastMessage:      time.Now(),
			ReceivedBranches: make(map[string]time.Time),
			BranchMutex:      new(sync.Mutex),
		}

		l.udpPool[address.String()] = conn

		go conn.udpReader()
		go conn.branchJanitor()
		go l.readRequests(conn)
	}

	return conn
}

func (l *Listener) registerTCPConn(netConn net.Conn) {
	conn := &Conn{
		Transport:        "tcp",
		Listener:         l,
		Conn:             netConn,
		Address:          netConn.RemoteAddr(),
		UdpReceiver:      nil,
		Closed:           false,
		Locked:           false,
		WriteBuffer:      new(bytes.Buffer),
		ReadMessage:      make(chan interface{}),
		LastMessage:      time.Time{},
		ReceivedBranches: make(map[string]time.Time),
		BranchMutex:      new(sync.Mutex),
	}

	go conn.tcpReader()
	go conn.branchJanitor()
	go l.readRequests(conn)
}

func (l *Listener) readRequests(conn *Conn) {
	for {
		req, err := conn.readRequest()

		l.requestChannel <- requestPackage{
			conn: conn,
			req:  req,
			err:  err,
		}

		if err == io.EOF {
			return
		}
	}
}

func (l *Listener) udpJanitor() {
	for {
		time.Sleep(time.Second * 10)

		var markClose []*Conn
		l.udpPoolMutex.Lock()
		for _, conn := range l.udpPool {
			if time.Now().Sub(conn.LastMessage) > time.Second*30 {
				markClose = append(markClose, conn)
			}
		}
		l.udpPoolMutex.Unlock()

		for _, conn := range markClose {
			conn.Close()
		}
	}
}
