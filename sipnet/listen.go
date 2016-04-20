package sipnet

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"fmt"
)

var (
	// ErrClosed is returned if AcceptRequest is called on a closed listener.
	ErrClosed        = errors.New("sip: closed")
	ErrInvalidBranch = errors.New("sip: invalid branch")
)

type requestPackage struct {
	responseWriter *ResponseWriter
	req            *Request
	err            error
}

// Listener represents a TCP and UDP wrapper listener.
type Listener struct {
	tcpListener      net.Listener
	udpListener      *net.UDPConn
	closed           bool
	requestChannel   chan requestPackage
	receivedBranches map[string]time.Time
	branchMutex      *sync.Mutex
}

// Listen listens on an address (IP:port) on both TCP and UDP.
func Listen(addr string) (*Listener, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		tcpListener.Close()
		return nil, err
	}

	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		tcpListener.Close()
		return nil, err
	}

	listener := &Listener{
		tcpListener:      tcpListener,
		udpListener:      udpListener,
		closed:           false,
		requestChannel:   make(chan requestPackage),
		receivedBranches: make(map[string]time.Time),
		branchMutex:      new(sync.Mutex),
	}

	go branchJanitor(listener)
	go handleTCPListening(listener)
	go handleUDPListening(listener)

	return listener, nil
}

func branchJanitor(listener *Listener) {
	for {
		time.Sleep(time.Second * 10)
		if listener.closed {
			return
		}

		listener.branchMutex.Lock()
		for branch, t := range listener.receivedBranches {
			if time.Now().Sub(t) > 30*time.Second {
				delete(listener.receivedBranches, branch)
			}
		}
		listener.branchMutex.Unlock()
	}
}

func handleTCPListening(listener *Listener) {
	for {
		conn, err := listener.tcpListener.Accept()
		if err != nil {
			if listener.closed {
				return
			}
			listener.requestChannel <- requestPackage{
				responseWriter: nil,
				req:            nil,
				err:            err,
			}
			continue
		}

		go handleTCPConn(listener, conn)
	}
}

func handleUDPListening(listener *Listener) {
	for {
		data := make([]byte, 65535)
		n, addr, err := listener.udpListener.ReadFrom(data)
		if err != nil {
			if listener.closed {
				return
			}

			listener.requestChannel <- requestPackage{
				responseWriter: nil,
				req:            nil,
				err:            err,
			}

			continue
		}

		if bytes.Compare(data[:n], []byte("\r\n\r\n")) == 0 {
			// Acknowledge keep alive
			listener.udpListener.WriteTo([]byte("\r\n"), addr)
			continue
		}

		req, err := ReadRequest(bytes.NewBuffer(data[:n]))

		listener.requestChannel <- requestPackage{
			responseWriter: NewResponseWriter(listener.udpListener, addr),
			req:            req,
			err:            err,
		}

		if err != nil {
			fmt.Println("read request error:", err)
		}
	}
}

func handleTCPConn(l *Listener, conn net.Conn) {
	defer conn.Close()

	for {
		req, err := ReadRequest(conn)
		if l.closed {
			return
		}

		if err == io.EOF {
			return
		}

		l.requestChannel <- requestPackage{
			responseWriter: NewResponseWriter(conn, conn.RemoteAddr()),
			req:            req,
			err:            err,
		}

		if err != nil {
			fmt.Println("read request error:", err)
		}
	}
}

// AcceptRequest blocks until it receives a Request message on either TCP or UDP
// listeners. Responses are to be written to *ResponseWriter. The request
// acceptor should be written in a stateless manner.
func (l *Listener) AcceptRequest() (*Request, *ResponseWriter, error) {
	for {
		if l.closed {
			return nil, nil, ErrClosed
		}
		resp := <-l.requestChannel

		if resp.err == nil {
			via, err := ParseVia(resp.req.Header.Get("Via"))
			if err != nil {
				return resp.req, resp.responseWriter, err
			}

			branch := via.Arguments.Get("branch")
			if branch == "" || len(branch) < 8 || branch[:7] != "z9hG4bK" {
				return resp.req, resp.responseWriter, ErrInvalidBranch
			}

			l.branchMutex.Lock()
			if _, found := l.receivedBranches[branch]; found {
				// Repeated message, ignore.
				l.branchMutex.Unlock()
				continue
			}

			l.receivedBranches[branch] = time.Now()
			l.branchMutex.Unlock()
		}

		return resp.req, resp.responseWriter, resp.err
	}
}

// Close closes both TCP and UDP listeners, and returns
func (l *Listener) Close() error {
	l.closed = true
	err := l.tcpListener.Close()
	if err != nil {
		l.udpListener.Close()
	} else {
		err = l.udpListener.Close()
	}

closeLoop:
	for {
		select {
		case <-l.requestChannel:
		default:
			break closeLoop
		}
	}

	return err
}

// Addr returns the address the listener is listening on.
func (l *Listener) Addr() net.Addr {
	return l.tcpListener.Addr()
}
