package sipnet

import (
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

// WriteTo writes the request data to a Conn. It automatically adds a
// a Content-Length to the header, calls Flush() on the Conn.
func (r *Request) WriteTo(conn *Conn) error {
	_, err := conn.Write([]byte(r.Method + " " + r.Server + " " + SIPVersion + "\r\n"))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))

	_, err = r.Header.WriteTo(conn)
	if err != nil {
		return err
	}

	_, err = conn.Write(r.Body)
	if err != nil {
		return err
	}

	return conn.Flush()
}
