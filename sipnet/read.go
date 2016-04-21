package sipnet

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// ErrBadMessage is returned by ReadRequest and ReadResponse if the message
// received failed to be parsed.
var ErrBadMessage = errors.New("sip: bad message")

// ReadRequest reads a SIP request (i.e. message from a UAC) from a reader.
func ReadRequest(rd io.Reader) (*Request, error) {
	buf := bufio.NewReader(rd)
	line, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if line[len(line)-2:] != "\r\n" {
		return nil, ErrBadMessage
	}

	args := strings.Split(line, " ")
	if len(args) != 3 {
		return nil, ErrBadMessage
	}

	r := NewRequest()
	r.Method = args[0]
	r.Server = args[1]
	r.SIPVersion = args[2][:len(args[2])-2]

	err = parseHeader(buf, r.Header)
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		return r, nil
	}

	body := make([]byte, length)
	_, err = buf.Read(body)
	if err != nil {
		return r, err
	}

	r.Body = body

	return r, nil
}

// ReadResponse reads a SIP response (i.e. message from a UAS) from a reader.
func ReadResponse(rd io.Reader) (*Response, error) {
	buf := bufio.NewReader(rd)
	line, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if line[len(line)-2:] != "\r\n" {
		return nil, ErrBadMessage
	}

	args := strings.Split(line, " ")
	if len(args) < 3 {
		return nil, ErrBadMessage
	}

	r := NewResponse()
	r.SIPVersion = args[0]
	r.StatusCode, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, err
	}

	r.Status = StatusText(r.StatusCode)

	err = parseHeader(buf, r.Header)
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		return r, nil
	}

	body := make([]byte, length)
	_, err = buf.Read(body)
	if err != nil {
		return r, err
	}

	r.Body = body

	return r, nil
}

func parseHeader(buf *bufio.Reader, h Header) error {
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return err
		}

		if line == "\r\n" {
			return nil
		}

		keyPosition := strings.Index(line, ":")
		if keyPosition == -1 {
			return ErrBadMessage
		}

		key := normalizeKey(strings.TrimSpace(line[:keyPosition]))
		value := strings.TrimSpace(line[keyPosition+1:])
		h.Set(key, value)
	}
}
