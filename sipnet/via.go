package sipnet

import (
	"regexp"
	"strings"
)

var viaRegexp = regexp.MustCompile("^(SIP\\/[^\\/]+)\\/([^ ]+) ([^;]+)(.+)$")

// Via represents the contents of the Via header line.
type Via struct {
	SIPVersion string
	Transport  string
	Client     string
	Arguments  HeaderArgs
}

// ParseVia parses a given Via header value into a Via.
func ParseVia(str string) (Via, error) {
	result := viaRegexp.FindStringSubmatch(str)
	if len(result) == 0 {
		return Via{}, ErrParseError
	}

	return Via{
		SIPVersion: strings.TrimSpace(result[1]),
		Transport:  strings.TrimSpace(result[2]),
		Client:     strings.TrimSpace(result[3]),
		Arguments:  ParseHeaderArgs(strings.TrimSpace(result[4])),
	}, nil
}

// String returns the string representation of the Via header line.
func (v Via) String() string {
	return v.SIPVersion + "/" + v.Transport + " " + v.Client +
		v.Arguments.SemicolonString()
}
