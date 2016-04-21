package sipnet

import (
	"bytes"
	"io"
	"net"
	"sync"
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

	locked       bool
	lockWait     *sync.WaitGroup
	lockedReader io.Reader

	writeBuffer *bytes.Buffer
	readMessage chan interface{}
	lastMessage time.Time
}

// Read reads raw bytes from the connection after it is locked.
func (c *Conn) Read(b []byte) (int, error) {
	if c.closed {
		return io.EOF
	}

	if c.lockedReader != nil {
		return c.lockedReader.Read(b)
	}

	msg, more := <-c.readMessage
	if !more {
		return io.EOF
	}

	if c.transport == "udp" {
		resp := msg.([]byte)
		buf := new(bytes.Buffer)
		buf.Write(resp)
		c.lockedReader = buf
		return buf.Read(b)
	}

	rd := msg.(io.Reader)
	c.lockedReader = rd
	return c.lockedReader.Read(b)
}

// Lock must be called to use Read(). It locks the connection to be read by
// the user.
func (c *Conn) Lock() {
	if !c.locked {
		c.locked = true
		c.lockWait.Add(1)
	}
}

// Unlock should be called after the user is finished reading custom
// data to the connection.
func (c *Conn) Unlock() {
	if c.locked {
		c.locked = false
		c.lockedReader = nil
		c.lockWait.Done()
	}
}

func (c *Conn) readRequest() (*Request, error) {
	for {
		if c.closed {
			return nil, io.EOF
		}

		for c.locked {
			c.lockWait.Wait()
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

		if c.locked {
			if c.lockedReader != nil {
				buf := c.lockedReader.(*bytes.Buffer)
				buf.Write(received)
				continue
			}

			c.readMessage <- received
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

		if c.locked {
			c.readMessage <- rd
			c.lockWait.Wait()
			continue
		}

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
