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

// WriteTo writes the response data to a ResponseWriter. It automatically adds a
// a Content-Length, CSeq, Call-ID and Via header. It also sets the Status message
// appropriately and automatically calls Flush() on the ResponseWriter.
func (r *Response) WriteTo(w *ResponseWriter, req *Request) error {
	_, err := w.Write([]byte(SIPVersion + " " + strconv.Itoa(r.StatusCode) +
		" " + StatusText(r.StatusCode) + "\r\n"))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))
	reqVia, err := ParseVia(req.Header.Get("Via"))
	if err != nil {
		return err
	}

	ipPort := strings.Split(w.Addr().String(), ":")
	reqVia.Arguments.Set("received", ipPort[0])
	reqVia.Arguments.Set("rport", ipPort[1])
	r.Header.Set("Via", reqVia.String())
	r.Header.Set("CSeq", req.Header.Get("CSeq"))
	r.Header.Set("Call-ID", req.Header.Get("Call-ID"))

	_, err = r.Header.WriteTo(w)
	if err != nil {
		return err
	}

	w.Write(r.Body)
	return w.Flush()
}

// BadRequest responds to a ResponseWriter with a StatusBadRequest for convenience.
func (r *Response) BadRequest(w *ResponseWriter, req *Request, reason string) {
	r.StatusCode = StatusBadRequest
	r.Header.Set("Reason-Phrase", reason)
	r.WriteTo(w, req)
}

// ServerError responds to a ResponseWriter with a StatusServerInternalError
// for convenience.
func (r *Response) ServerError(w *ResponseWriter, req *Request, reason string) {
	r.StatusCode = StatusServerInternalError
	r.Header.Set("Reason-Phrase", reason)
	r.WriteTo(w, req)
}
