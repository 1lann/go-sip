package server

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/1lann/go-sip/sipnet"
	"strings"
	"sync"
	"time"
)

// TODO: Place this in a configuration file
var hostname = "localhost"

type authSession struct {
	nonce   string
	user    sipnet.User
	conn    *sipnet.Conn
	created time.Time
}

// a map[call id]authSession pair
var authSessions = make(map[string]authSession)
var authSessionMutex = new(sync.Mutex)

var ErrInvalidAuthHeader = errors.New("server: invalid authentication header")

func generateNonce(size int) string {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func parseAuthHeader(header string) (sipnet.HeaderArgs, error) {
	if len(header) < 8 || strings.ToLower(header[:7]) != "digest " {
		return nil, ErrInvalidAuthHeader
	}

	return sipnet.ParsePairs(header[7:]), nil
}

func requestAuthentication(r *sipnet.Request, conn *sipnet.Conn, from sipnet.User) {
	resp := sipnet.NewResponse()

	callId := r.Header.Get("Call-ID")
	if callId == "" {
		resp.BadRequest(conn, r, "Missing required Call-ID header.")
		return
	}

	nonce := generateNonce(32)

	resp.StatusCode = sipnet.StatusUnauthorized
	// No auth header, deny.
	resp.Header.Set("From", from.String())
	from.Arguments.Del("tag")
	resp.Header.Set("To", from.String())

	authArgs := make(sipnet.HeaderArgs)
	authArgs.Set("realm", hostname)
	authArgs.Set("qop", "auth")
	authArgs.Set("nonce", nonce)
	authArgs.Set("opaque", "")
	authArgs.Set("stale", "FALSE")
	authArgs.Set("algorithm", "MD5")
	resp.Header.Set("WWW-Authenticate", "Digest "+authArgs.CommaString())

	authSessionMutex.Lock()
	authSessions[callId] = authSession{
		nonce:   nonce,
		user:    from,
		conn:    conn,
		created: time.Now(),
	}
	authSessionMutex.Unlock()

	resp.WriteTo(conn, r)
	return
}

func md5Hex(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func checkAuthorization(r *sipnet.Request, conn *sipnet.Conn,
	authArgs sipnet.HeaderArgs, user sipnet.User) {
	callId := r.Header.Get("Call-ID")
	authSessionMutex.Lock()
	session, found := authSessions[callId]
	authSessionMutex.Unlock()
	if !found {
		requestAuthentication(r, conn, user)
		return
	}

	if authArgs.Get("username") != user.URI.Username {
		requestAuthentication(r, conn, user)
		return
	}

	if authArgs.Get("nonce") != session.nonce {
		requestAuthentication(r, conn, user)
		return
	}

	username := user.URI.Username
	account, found := accounts[username]
	if !found {
		requestAuthentication(r, conn, user)
		return
	}

	ha1 := md5Hex(username + ":" + hostname + ":" + account.password)
	ha2 := md5Hex(sipnet.MethodRegister + ":" + authArgs.Get("uri"))
	response := md5Hex(ha1 + ":" + session.nonce + ":" + authArgs.Get("nc") +
		":" + authArgs.Get("cnonce") + ":auth:" + ha2)

	if response != authArgs.Get("response") {
		requestAuthentication(r, conn, user)
		return
	}

	if r.Header.Get("Expires") == "0" {
		registeredUsersMutex.Lock()
		delete(registeredUsers, username)
		registeredUsersMutex.Unlock()
		println("logged out " + username)
	} else {
		registerUser(session)
		println("registered " + username)
	}

	resp := sipnet.NewResponse()
	resp.StatusCode = sipnet.StatusOK
	resp.Header.Set("From", user.String())

	user.Arguments.Set("tag", generateNonce(5))
	resp.Header.Set("To", user.String())
	resp.WriteTo(conn, r)

	return
}

// HandleRegister handles REGISTER SIP requests.
func HandleRegister(r *sipnet.Request, conn *sipnet.Conn) {
	from, to, err := sipnet.ParseUserHeader(r.Header)
	if err != nil {
		resp := sipnet.NewResponse()
		resp.BadRequest(conn, r, "Failed to parse From or To header.")
		return
	}

	if to.URI.UserDomain() != from.URI.UserDomain() {
		resp := sipnet.NewResponse()
		resp.BadRequest(conn, r, "User in To and From fields do not match.")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		requestAuthentication(r, conn, from)
		return
	}

	args, err := parseAuthHeader(authHeader)
	if err != nil {
		resp := sipnet.NewResponse()
		resp.BadRequest(conn, r, "Failed to parse Authorization header.")
		return
	}

	checkAuthorization(r, conn, args, from)
}

func registrationJanitor() {
	for {
		authSessionMutex.Lock()
		for callId, session := range authSessions {
			if time.Now().Sub(session.created) > time.Second*30 {
				delete(authSessions, callId)
			}
		}
		authSessionMutex.Unlock()
		time.Sleep(time.Second * 10)
	}
}

func init() {
	go registrationJanitor()
}
