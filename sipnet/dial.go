// Package sipnet contains tools for connecting and communicating to SIP UAs.
package sipnet

import (
	"errors"
	"net"
	"time"
)

// ErrInvalidTransport is returned by Dial if the transport provided is
// invalid.
var ErrInvalidTransport = errors.New("sip: invalid transport")

// Dial creates a connection to a SIP UA. It does NOT "dial" a
// user. addr is an IP:port string,
// transport is the transport protocol to be used, (i.e. "tcp" or "udp").
//
// After dialling, you should use ReadResponse to read from the connection,
// and Request.WriteTo to write requests to the connection.
func Dial(addr, transport string) (net.Conn, error) {
	if transport == "tcp" {
		conn, err := net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			return nil, err
		}

		return conn, nil
	} else if transport == "udp" {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			return nil, err
		}

		return conn, nil
	} else {
		return nil, ErrInvalidTransport
	}
}
