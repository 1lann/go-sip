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
			transport:    "udp",
			listener:     l,
			conn:         l.udpListener,
			address:      address,
			udpReceiver:  make(chan []byte),
			closed:       false,
			locked:       false,
			lockWait:     new(sync.WaitGroup),
			lockedReader: nil,
			writeBuffer:  new(bytes.Buffer),
			readMessage:  make(chan interface{}),
			lastMessage:  time.Now(),
		}

		l.udpPool[address.String()] = conn

		go l.readRequests(conn)
	}

	return conn
}

func (l *Listener) registerTCPConn(netConn net.Conn) {
	conn := &Conn{
		transport:    "tcp",
		listener:     l,
		conn:         netConn,
		address:      netConn.RemoteAddr(),
		udpReceiver:  nil,
		closed:       false,
		locked:       false,
		lockWait:     new(sync.WaitGroup),
		lockedReader: nil,
		writeBuffer:  new(bytes.Buffer),
		readMessage:  make(chan interface{}),
		lastMessage:  time.Time{},
	}

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
	l.udpPoolMutex.Lock()
	for address, conn := range l.udpPool {
		if time.Now().Sub(conn.lastMessage) > time.Second*30 {
			// Disconnect
			conn.Close()
			delete(l.udpPool, address)
		}
	}
	l.udpPoolMutex.Unlock()
}
