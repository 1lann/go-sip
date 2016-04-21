package server

import (
	"github.com/1lann/go-sip/sipnet"
	"sync"
)

type registeredUser struct {
	username string
	conn     *sipnet.Conn
}

var registeredUsers = make(map[string]registeredUser)
var registeredUsersMutex = new(sync.Mutex)

func registerUser(session authSession) {
	registeredUsersMutex.Lock()
	defer registeredUsersMutex.Unlock()

	username := session.user.URI.Username
	if connected, found := registeredUsers[username]; found {
		connected.conn.Close()
	}

	newUser := registeredUser{
		username: username,
		conn:     session.conn,
	}

	registeredUsers[username] = newUser
}
