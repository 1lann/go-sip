package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/1lann/go-sip/sipnet"
)

var errGiveUp = errors.New("server: give up")
var errUnknownResponse = errors.New("server: unknown response")

// HandleInvite handles INVITE SIP requests and attempts to make a call.
func HandleInvite(r *sipnet.Request, conn *sipnet.Conn) {
	from, to, err := sipnet.ParseUserHeader(r.Header)
	if err != nil {
		resp := sipnet.NewResponse()
		resp.BadRequest(conn, r, "Failed to parse From or To header.")
		return
	}

	username := from.URI.Username

	registeredUsersMutex.Lock()
	user, found := registeredUsers[username]
	registeredUsersMutex.Unlock()

	if !found || user.conn != conn {
		resp := sipnet.NewResponse()
		resp.StatusCode = sipnet.StatusForbidden
		resp.Header.Set("Reason-Phrase", "Not registered.")
		resp.WriteTo(conn, r)
		return
	}

	recipient := to.URI.Username
	recipientUser, found := registeredUsers[recipient]
	if !found {
		resp := sipnet.NewResponse()
		resp.StatusCode = sipnet.StatusNotFound
		resp.WriteTo(conn, r)
		return
	}

	recipientUser.conn.Lock()
	defer recipientUser.conn.Unlock()
	conn.Lock()
	defer conn.Unlock()

	fmt.Println("calling " + recipientUser.username)

	initiateCall(r, conn, recipientUser.conn)
}

func initiateCall(initialRequest *sipnet.Request,
	from *sipnet.Conn, to *sipnet.Conn) {
	from.Lock()
	defer from.Unlock()
	to.Lock()
	defer to.Unlock()

	trying(initialRequest, from)
	initialRequest.WriteTo(to)

	wg := new(sync.WaitGroup)

	wg.Add(2)

	go func() {
		defer wg.Done()
		var lastRequest *sipnet.Request
		for {
			read := from.Read()
			switch read.(type) {
			case *sipnet.Request:
				req := read.(*sipnet.Request)
				fmt.Println("from --> to request, forwarding")
				fmt.Println(req)
				lastRequest = req
				req.WriteTo(to)
			case *sipnet.Response:
				resp := read.(*sipnet.Response)
				fmt.Println("from --> to response, forwarding")
				fmt.Println(resp)

				resp.WriteTo(to, lastRequest)
			case error:
				err := read.(error)
				fmt.Println("TODO: from error:", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		lastRequest := initialRequest
		for {
			read := to.Read()
			switch read.(type) {
			case *sipnet.Request:
				req := read.(*sipnet.Request)

				if req.Method == sipnet.MethodOptions {
					fmt.Println("responding with options")
					resp := sipnet.NewResponse()
					resp.StatusCode = sipnet.StatusOK
					resp.Header.Set("Allow", "INVITE, ACK, CANCEL, OPTIONS, BYE")
					resp.Header.Set("Accept", "application/sdp")
					resp.Header.Set("Accept-Encoding", "gzip")
					resp.Header.Set("Accept-Language", "en")
					resp.Header.Set("Content-Type", "application/sdp")
					resp.Body = initialRequest.Body
					resp.WriteTo(from, req)
					break
				}

				fmt.Println("to --> from request, forwarding")
				fmt.Println(req)
				lastRequest = req

				req.WriteTo(from)
			case *sipnet.Response:
				resp := read.(*sipnet.Response)
				if resp.StatusCode == sipnet.StatusTrying {
					fmt.Println("stop trying :P")
					break
				}

				fmt.Println("to --> from response, forwarding")
				fmt.Println(resp)

				resp.WriteTo(from, lastRequest)
			case error:
				err := read.(error)
				fmt.Println("TODO: from error:", err)
				return
			}
		}
	}()

	wg.Wait()
}

func trying(r *sipnet.Request, conn *sipnet.Conn) {
	resp := sipnet.NewResponse()
	resp.StatusCode = sipnet.StatusTrying
	resp.WriteTo(conn, r)
}

func waitResponse(conn *sipnet.Conn) (*sipnet.Response, error) {
	resp := conn.Read()
	switch resp.(type) {
	case *sipnet.Response:
		return resp.(*sipnet.Response), nil
	case error:
		return nil, resp.(error)
	default:
		return nil, errUnknownResponse
	}
}

func makeUnreliableRequest(r *sipnet.Request, fromConn *sipnet.Conn,
	toConn *sipnet.Conn) *sipnet.Response {
	for {
		receivedResponse := false
		responseChannel := make(chan interface{})

		go func() {
			for i := 0; i < 10; i++ {
				if receivedResponse {
					return
				}

				err := r.WriteTo(toConn)
				if err != nil {
					fmt.Println("write error:", err)
					responseChannel <- err
				}
				time.Sleep(time.Millisecond * 500)
			}
			responseChannel <- errGiveUp
		}()

		go func() {
			responseChannel <- toConn.Read()
		}()

		resp := <-responseChannel
		receivedResponse = true

		switch resp.(type) {
		case *sipnet.Response:
			return resp.(*sipnet.Response)
		case *sipnet.Request:
			req := resp.(*sipnet.Request)

			resp := sipnet.NewResponse()
			resp.StatusCode = sipnet.StatusOK
			resp.Header.Set("Allow", "INVITE, ACK, CANCEL, OPTIONS, BYE")
			resp.Header.Set("Accept", "application/sdp")
			resp.Header.Set("Accept-Encoding", "gzip")
			resp.Header.Set("Accept-Language", "en")
			resp.Header.Set("Content-Type", "application/sdp")
			resp.Body = r.Body
			resp.WriteTo(fromConn, req)

			break
		case error:
			fmt.Println("received error response:", resp.(error))
			return nil
		default:
			fmt.Println("received unknown response:", resp)
			return nil
		}
	}
}
