package sipnet

import (
	"bytes"
	"net"
	"strconv"
)

// SIPVersion is the version of SIP used by this library.
const SIPVersion = "SIP/2.0"

// SIP request methods.
const (
	MethodInvite   = "INVITE"
	MethodAck      = "ACK"
	MethodBye      = "BYE"
	MethodCancel   = "CANCEL"
	MethodRegister = "REGISTER"
	MethodOptions  = "OPTIONS"
	MethodInfo     = "INFO"
)

// Request represents a SIP request (i.e. a message sent by a UAC to a UAS).
type Request struct {
	Method     string
	Server     string
	SIPVersion string
	Header     Header
	Body       []byte
}

// NewRequest returns a new request.
func NewRequest() *Request {
	return &Request{
		SIPVersion: SIPVersion,
		Header:     make(Header),
	}
}

// Flushable is used by Request.WriteTo to determine whether or not the
// provided connection is flushable, and if so, writes and then flushes it.
type Flushable interface {
	Flush() error
}

// WriteTo writes the request data to a Conn. It automatically adds a
// a Content-Length to the header, calls Flush() on the Conn.
func (r *Request) WriteTo(conn net.Conn) error {
	buf := new(bytes.Buffer)

	buf.Write([]byte(r.Method + " " + r.Server + " " + SIPVersion + "\r\n"))

	r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))

	r.Header.WriteTo(buf)
	buf.Write(r.Body)

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if flushConn, ok := conn.(Flushable); ok {
		return flushConn.Flush()
	}

	return nil
}
