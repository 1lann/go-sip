package sipnet

import (
	"bytes"
	"io"
	"net"
)

// ResponseWriter is a writer that is used to respond to a Request from an
// AcceptRequest.
type ResponseWriter struct {
	addr      net.Addr
	transport string
	buffer    *bytes.Buffer
	writer    io.Writer
}

// Write writes into the buffer of the ResponseWriter. Once the whole response
// is ready to be sent, you must call Flush.
func (r *ResponseWriter) Write(b []byte) (int, error) {
	return r.buffer.Write(b)
}

// Flush flushes the data written to the buffer of a ResponseWriter into
// the underlying writer (i.e. the network connection).
func (r *ResponseWriter) Flush() error {
	if r.transport == "udp" {
		udpConn := r.writer.(*net.UDPConn)
		_, err := udpConn.WriteTo(r.buffer.Bytes(), r.addr)
		return err
	}

	_, err := r.buffer.WriteTo(r.writer)
	return err
}

// Addr returns the IP:port address of where the response will be written to.
func (r *ResponseWriter) Addr() net.Addr {
	return r.addr
}

// Transport returns the type of transport is used in the underlying writer
// (i.e. either "tcp" or "udp").
func (r *ResponseWriter) Transport() string {
	return r.transport
}

func NewResponseWriter(w io.Writer, addr net.Addr) *ResponseWriter {
	transport := "tcp"

	_, isUDP := w.(*net.UDPConn)
	if isUDP {
		transport = "udp"
	}

	return &ResponseWriter{
		addr:      addr,
		transport: transport,
		buffer:    new(bytes.Buffer),
		writer:    w,
	}
}
