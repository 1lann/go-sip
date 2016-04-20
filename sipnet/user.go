package sipnet

import (
	"errors"
	"regexp"
	"strings"
)

// ErrParseError is returned when a piece of data fails to be parsed
// (i.e. a user line or via).
var ErrParseError = errors.New("sip: parse error")

var nameRegexp = regexp.MustCompile("^([^<]*)<([^>]+)>(.*)$")

// User represents a SIP user.
type User struct {
	Name      string
	URI       URI
	Arguments HeaderArgs
}

// String returns the string representation of a user to be used on user lines.
func (u User) String() string {
	if u.Name == "" {
		return "<" + u.URI.String() + ">" + u.Arguments.SemicolonString()
	}
	return u.Name + " <" + u.URI.String() + ">" +
		u.Arguments.SemicolonString()
}

// ParseUser parses a given user line into a User.
func ParseUser(str string) (User, error) {
	result := nameRegexp.FindStringSubmatch(str)
	if len(result) == 0 {
		uri, err := ParseURI(str)
		if err != nil {
			return User{}, err
		}

		return User{
			URI:       uri,
			Arguments: make(HeaderArgs),
		}, nil
	}

	uri, err := ParseURI(strings.TrimSpace(result[2]))
	if err != nil {
		return User{}, err
	}

	return User{
		Name:      strings.TrimSpace(result[1]),
		URI:       uri,
		Arguments: ParseHeaderArgs(strings.TrimSpace(result[3])),
	}, nil
}

// ParseHeader returns the users from the To, and the From fields respectively
// from the header.
func ParseUserHeader(h Header) (User, User, error) {
	var from User
	to, err := ParseUser(h.Get("To"))
	if err != nil {
		from, _ = ParseUser(h.Get("From"))
	} else {
		from, err = ParseUser(h.Get("From"))
	}

	return to, from, err
}
