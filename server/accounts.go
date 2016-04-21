// Package server contains a SIP server which can manage authentication
// and signalling between multiple clients.
package server

type account struct {
	password string
}

var accounts = make(map[string]account)

func init() {
	accounts["jason"] = account{"password"}
	accounts["phone"] = account{"password"}
	accounts["1012"] = account{"password"}
	accounts["1011"] = account{"1234"}
}
