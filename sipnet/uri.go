package sipnet

import (
	"regexp"
	"strings"
)

var uriRegexp = regexp.MustCompile("^([A-Za-z]+):([^@]+)@([^\\s;]+)(.*)$")

// URI represents a Uniform Resource Identifier.
type URI struct {
	Scheme    string
	Username  string
	Domain    string
	Arguments HeaderArgs
}

// ParseURI parses a given URI into a URI struct.
func ParseURI(str string) (URI, error) {
	result := uriRegexp.FindStringSubmatch(str)
	if len(result) == 0 {
		return URI{}, ErrParseError
	}

	arguments := make(HeaderArgs)
	if result[4] != "" && result[4][0] == ';' {
		arguments = ParseHeaderArgs(strings.TrimSpace(result[4][1:]))
	}

	return URI{
		Scheme:    result[1],
		Username:  result[2],
		Domain:    result[3],
		Arguments: arguments,
	}, nil
}

// String returns the full text representation of the URI with additional
// semicolon arguments.
func (u URI) String() string {
	return u.SchemeUserDomain() + u.Arguments.SemicolonString()
}

// SchemeUserDomain returns the text representation of the scheme:user@domain.
func (u URI) SchemeUserDomain() string {
	return u.Scheme + ":" + u.UserDomain()
}

// UserDomain returns the text representation of user@domain.
func (u URI) UserDomain() string {
	return u.Username + "@" + u.Domain
}
