package server

type account struct {
	password string
}

var accounts = make(map[string]account)

func init() {
	accounts["jason"] = account{"password"}
}
