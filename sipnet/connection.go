package sipnet

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"

	"fmt"
)

// Conn represents a connection with a UA. It can be on UDP or TCP.
type Conn struct {
	Transport   string
	Listener    *Listener
	Conn        net.Conn
	Address     net.Addr
	UdpReceiver chan []byte
	Closed      bool
	Locked      bool
	WriteBuffer *bytes.Buffer
	ReadMessage chan interface{}
	LastMessage time.Time

	ReceivedBranches map[string]time.Time
	BranchMutex      *sync.Mutex
}

// Read reads either a *Request, a *Response, or an error from the connection.
func (c *Conn) Read() interface{} {
	if c.Closed {
		return io.EOF
	}

	msg, more := <-c.ReadMessage
	if !more {
		return io.EOF
	}

	return msg
}

// Lock must be called to use Read(). It locks the connection to be read by
// the user rather than by read by AcceptRequest().
func (c *Conn) Lock() {
	c.Locked = true
}

// Unlock should be called after the user is finished reading custom
// data to the connection.
func (c *Conn) Unlock() {
	c.Locked = false
}

func (c *Conn) readRequest() (*Request, error) {
	for {
		if c.Closed {
			return nil, io.EOF
		}

		for c.Locked {
			time.Sleep(time.Second * 2)
		}

		msg, more := <-c.ReadMessage
		if !more {
			return nil, io.EOF
		}

		if c.Locked {
			c.ReadMessage <- msg
			continue
		}

		switch msg.(type) {
		case error:
			return nil, msg.(error)
		case *Request:
			return msg.(*Request), nil
		default:
			fmt.Println("warning: unhandled message type (likely a response)")
		}
	}
}

func (c *Conn) udpReader() {
	for {
		received, more := <-c.UdpReceiver
		if !more {
			return
		}

		c.LastMessage = time.Now()
		if bytes.Compare(received, []byte("\r\n\r\n")) == 0 {
			// Acknowledge keep alive
			c.Write([]byte("\r\n"))
			continue
		}

		rd := bytes.NewReader(received)
		if bytes.Compare(received[:3], []byte("SIP")) == 0 {
			resp, err := ReadResponse(rd)
			if err != nil {
				c.ReadMessage <- err
				continue
			}
			c.ReadMessage <- resp
			continue
		}

		req, err := ReadRequest(rd)
		if err != nil {
			c.ReadMessage <- err
			continue
		}

		c.ReadMessage <- req
	}
}

func (c *Conn) tcpReader() {
	for {
		buf := make([]byte, 3)
		_, err := io.ReadFull(c.Conn, buf)
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				c.Close()
				return
			}
		}

		rd := io.MultiReader(bytes.NewReader(buf), c.Conn)

		if bytes.Compare(buf, []byte("SIP")) == 0 {
			resp, err := ReadResponse(rd)
			if err != nil {
				c.ReadMessage <- err
				continue
			}
			c.ReadMessage <- resp
			continue
		}

		req, err := ReadRequest(rd)
		if err != nil {
			c.ReadMessage <- err
			continue
		}

		c.ReadMessage <- req
	}
}

func (c *Conn) writeReceivedUDP(b []byte) {
	if c.Closed {
		return
	}

	c.UdpReceiver <- b
}

// Write writes data to a buffer.
func (c *Conn) Write(b []byte) (int, error) {
	if c.Closed {
		return 0, io.ErrClosedPipe
	}

	return c.WriteBuffer.Write(b)
}

// Flush flushes the buffered data to be written. In the case of using UDP,
// the buffered data will be written in a single UDP packet.
func (c *Conn) Flush() error {
	if c.Closed {
		return io.ErrClosedPipe
	}

	if c.Transport == "udp" {
		udpConn := c.Conn.(*net.UDPConn)
		_, err := udpConn.WriteTo(c.WriteBuffer.Bytes(), c.Address)
		c.WriteBuffer.Reset()
		return err
	}

	_, err := c.Conn.Write(c.WriteBuffer.Bytes())
	c.WriteBuffer.Reset()

	return err
}

// Addr returns the network address of the connected UA.
func (c *Conn) Addr() net.Addr {
	return c.Address
}

// Close closes the connection.
func (c *Conn) Close() error {
	if c.Closed {
		return nil
	}

	c.Closed = true

	if c.Transport == "udp" {
		if c.Listener != nil {
			c.Listener.udpPoolMutex.Lock()
			delete(c.Listener.udpPool, c.Address.String())
			c.Listener.udpPoolMutex.Unlock()
		}
		close(c.UdpReceiver)
		return nil
	}

	return c.Conn.Close()
}

func (c *Conn) branchJanitor() {
	for {
		time.Sleep(time.Second * 10)
		if c.Closed {
			return
		}

		c.BranchMutex.Lock()
		for branch, t := range c.ReceivedBranches {
			if time.Now().Sub(t) > 30*time.Second {
				delete(c.ReceivedBranches, branch)
			}
		}
		c.BranchMutex.Unlock()
	}
}
