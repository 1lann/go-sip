package server

type registeredUser struct {
	username string
	address  string
	tag      string
}

var registeredUsers map[string]registeredUser

func registerUser(session authSession) (string, error) {

}
