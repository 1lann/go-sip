package sipnet

import (
	"io"
	"strconv"
)

// SIPVersion is the version of SIP used by this library.
const SIPVersion = "SIP/2.0"

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

// func (r *Request) SetAuthentication()

// WriteTo writes the request data to a writer. It automatically adds a
// a Content-Length to the header.
func (r *Request) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte(r.Method + " " + r.Server + SIPVersion + "\r\n"))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))

	_, err = r.Header.WriteTo(w)
	if err != nil {
		return err
	}

	_, err = w.Write(r.Body)

	return err
}
