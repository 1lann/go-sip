package sipnet

import (
	"bytes"
	"io"
	"net"
	"time"

	"fmt"
)

type Conn struct {
	transport   string
	listener    *Listener
	conn        net.Conn
	address     net.Addr
	udpReceiver chan []byte
	closed      bool
	locked      bool
	writeBuffer *bytes.Buffer
	readMessage chan interface{}
	lastMessage time.Time
}

// Read reads either a *Request, a *Response, or an error from the connection.
func (c *Conn) Read() interface{} {
	if c.closed {
		return io.EOF
	}

	msg, more := <-c.readMessage
	if !more {
		return io.EOF
	}

	return msg
}

// Lock must be called to use Read(). It locks the connection to be read by
// the user.
func (c *Conn) Lock() {
	c.locked = true
}

// Unlock should be called after the user is finished reading custom
// data to the connection.
func (c *Conn) Unlock() {
	c.locked = false
}

func (c *Conn) readRequest() (*Request, error) {
	for {
		if c.closed {
			return nil, io.EOF
		}

		for c.locked {
			time.Sleep(time.Second * 2)
		}

		msg, more := <-c.readMessage
		if !more {
			return nil, io.EOF
		}

		if c.locked {
			c.readMessage <- msg
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
		received, more := <-c.udpReceiver
		if !more {
			return
		}

		c.lastMessage = time.Now()
		if bytes.Compare(received, []byte("\r\n\r\n")) == 0 {
			// Acknowledge keep alive
			c.Write([]byte("\r\n"))
			continue
		}

		rd := bytes.NewReader(received)
		if bytes.Compare(received[:3], []byte("SIP")) == 0 {
			resp, err := ReadResponse(rd)
			if err != nil {
				c.readMessage <- err
				continue
			}
			c.readMessage <- resp
			continue
		}

		req, err := ReadRequest(rd)
		if err != nil {
			c.readMessage <- err
			continue
		}
		c.readMessage <- req
	}
}

func (c *Conn) tcpReader() {
	for {
		buf := make([]byte, 3)
		_, err := io.ReadFull(c.conn, buf)
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				c.Close()
				return
			}
		}

		rd := io.MultiReader(bytes.NewReader(buf), c.conn)

		if bytes.Compare(buf, []byte("SIP")) == 0 {
			resp, err := ReadResponse(rd)
			if err != nil {
				c.readMessage <- err
				continue
			}
			c.readMessage <- resp
			continue
		}

		req, err := ReadRequest(rd)
		if err != nil {
			c.readMessage <- err
			continue
		}
		c.readMessage <- req
	}
}

func (c *Conn) writeReceivedUDP(b []byte) {
	if c.closed {
		return
	}

	c.udpReceiver <- b
}

func (c *Conn) Write(b []byte) (int, error) {
	if c.closed {
		return 0, io.ErrClosedPipe
	}

	return c.writeBuffer.Write(b)
}

func (c *Conn) Flush() error {
	if c.closed {
		return io.ErrClosedPipe
	}

	if c.transport == "udp" {
		udpConn := c.conn.(*net.UDPConn)
		_, err := udpConn.WriteTo(c.writeBuffer.Bytes(), c.address)
		c.writeBuffer.Reset()
		return err
	}

	_, err := c.conn.Write(c.writeBuffer.Bytes())
	c.writeBuffer.Reset()
	return err

}

func (c *Conn) Transport() string {
	return c.transport
}

func (c *Conn) Addr() net.Addr {
	return c.address
}

func (c *Conn) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true

	if c.transport == "udp" {
		if c.listener != nil {
			c.listener.udpPoolMutex.Lock()
			delete(c.listener.udpPool, c.address.String())
			c.listener.udpPoolMutex.Unlock()
		}
		close(c.udpReceiver)
		return nil
	}

	return c.conn.Close()
}
