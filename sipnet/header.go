package sipnet

import (
	"io"
	"strings"
)

// Header represents the headers of a SIP Request or Response.
type Header map[string]string

// Del deletes the key and its value from the header. Deleting a non-existent
// key is a no-op.
func (h Header) Del(key string) {
	delete(h, normalizeKey(key))
}

// Get returns the value at a given key. It returns an empty string if the
// key does not exist.
func (h Header) Get(key string) string {
	return h[normalizeKey(key)]
}

// Set sets a header key with a value.
func (h Header) Set(key, value string) {
	h[normalizeKey(key)] = value
}

// WriteTo writes the header data to a writer, with an additional CRLF
// (i.e. "\r\n") at the end.
func (h Header) WriteTo(w io.Writer) (int64, error) {
	var total int64
	for key, value := range h {
		n, err := w.Write([]byte(key + ": " + value + "\r\n"))
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	n, err := w.Write([]byte("\r\n"))
	total += int64(n)
	return total, err
}

func normalizeKey(key string) string {
	return strings.Title(strings.ToLower(key))
}
