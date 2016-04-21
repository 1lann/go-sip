package sipnet

import (
	"strconv"
	"strings"
)

// Response represents a SIP response (i.e. a message sent by a UAS to a UAC).
type Response struct {
	StatusCode int
	Status     string
	SIPVersion string
	Header     Header
	Body       []byte
}

// NewResponse returns a new response.
func NewResponse() *Response {
	return &Response{
		SIPVersion: SIPVersion,
		Header:     make(Header),
	}
}

// WriteTo writes the response data to a Conn. It automatically adds a
// a Content-Length, CSeq, Call-ID and Via header. It also sets the Status message
// appropriately and automatically calls Flush() on the Conn.
func (r *Response) WriteTo(conn *Conn, req *Request) error {
	_, err := conn.Write([]byte(SIPVersion + " " + strconv.Itoa(r.StatusCode) +
		" " + StatusText(r.StatusCode) + "\r\n"))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))
	reqVia, err := ParseVia(req.Header.Get("Via"))
	if err != nil {
		return err
	}

	ipPort := strings.Split(conn.Addr().String(), ":")
	reqVia.Arguments.Set("received", ipPort[0])
	reqVia.Arguments.Set("rport", ipPort[1])
	r.Header.Set("Via", reqVia.String())
	r.Header.Set("CSeq", req.Header.Get("CSeq"))
	r.Header.Set("Call-ID", req.Header.Get("Call-ID"))

	_, err = r.Header.WriteTo(conn)
	if err != nil {
		return err
	}

	conn.Write(r.Body)
	return conn.Flush()
}

// BadRequest responds to a Conn with a StatusBadRequest for convenience.
func (r *Response) BadRequest(conn *Conn, req *Request, reason string) {
	r.StatusCode = StatusBadRequest
	r.Header.Set("Reason-Phrase", reason)
	r.WriteTo(conn, req)
}

// ServerError responds to a Conn with a StatusServerInternalError
// for convenience.
func (r *Response) ServerError(conn *Conn, req *Request, reason string) {
	r.StatusCode = StatusServerInternalError
	r.Header.Set("Reason-Phrase", reason)
	r.WriteTo(conn, req)
}
