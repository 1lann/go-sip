package sipnet

import (
	"bytes"
	"strconv"
	"strings"
)

type HeaderArgs map[string]string

// ParseList parses a comma, semicolon, or new line seperated list of values
// and returns list elements.
//
// Lifted from https://code.google.com/p/gorilla/source/browse/http/parser/parser.go
// which was ported from urllib2.parse_http_list, from the Python
// standard library.
func ParseList(value string) []string {
	var list []string
	var escape, quote bool
	b := new(bytes.Buffer)
	for _, r := range value {
		switch {
		case escape:
			b.WriteRune(r)
			escape = false
		case quote:
			if r == '\\' {
				escape = true
			} else {
				if r == '"' {
					quote = false
				}
				b.WriteRune(r)
			}
		case r == ',' || r == ';' || r == '\n':
			list = append(list, strings.TrimSpace(b.String()))
			b.Reset()
		case r == '"':
			quote = true
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	// Append last part.
	if s := b.String(); s != "" {
		list = append(list, strings.TrimSpace(s))
	}
	return list
}

// ParsePairs extracts key/value pairs from comma, semicolon, or new line
// seperated values.
//
// Lifted from https://code.google.com/p/gorilla/source/browse/http/parser/parser.go
func ParsePairs(value string) HeaderArgs {
	m := make(HeaderArgs)
	for _, pair := range ParseList(strings.TrimSpace(value)) {
		if i := strings.Index(pair, "="); i < 0 {
			m[pair] = ""
		} else {
			v := pair[i+1:]
			if v[0] == '"' && v[len(v)-1] == '"' {
				v = v[1 : len(v)-1]
			}
			m[pair[:i]] = v
		}
	}
	return m
}

// Parse parses header arguments from a full header.
func ParseHeaderArgs(str string) HeaderArgs {
	argLocation := strings.Index(str, ";")
	if argLocation < 0 {
		return make(HeaderArgs)
	}

	return ParsePairs(str[argLocation+1:])
}

// Del deletes the key and its value from the header arguments. Deleting a non-existent
// key is a no-op.
func (h HeaderArgs) Del(key string) {
	delete(h, key)
}

// Get returns the value at a given key. It returns an empty string if the
// key does not exist.
func (h HeaderArgs) Get(key string) string {
	return h[key]
}

// Set sets a header argument key with a value.
func (h HeaderArgs) Set(key, value string) {
	h[key] = value
}

// SemicolonString returns the header arguments as a semicolon
// seperated unquoted strings with a leading semicolon.
func (h HeaderArgs) SemicolonString() string {
	var result string
	for key, value := range h {
		if value == "" {
			result += ";" + key
		} else {
			result += ";" + key + "=" + value
		}
	}
	return result
}

// CommaString returns the header arguments as a comma and space
// seperated string.
func (h HeaderArgs) CommaString() string {
	if len(h) == 0 {
		return ""
	}

	var result string
	for key, value := range h {
		result += key + "=" + strconv.Quote(value) + ", "
	}
	return result[:len(result)-2]
}

// CRLFString returns the header arguments as a CRLF seperated string.
func (h HeaderArgs) CRLFString() string {
	if len(h) == 0 {
		return ""
	}

	var result string
	for key, value := range h {
		result += key + "=" + value + "\r\n"
	}
	return result
}
