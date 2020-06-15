// Package janus is a Golang implementation of the Janus API, used to interact
// with the Janus WebRTC Gateway.
package janus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const httpTimeout = 30 * time.Second

var debug = false

func unexpected(request string) error {
	return fmt.Errorf("Unexpected response received to '%s' request", request)
}

func newRequest(method string) (map[string]interface{}, chan interface{}) {
	req := make(map[string]interface{}, 8)
	req["janus"] = method
	return req, make(chan interface{})
}

//IGatewaway ...
type IGatewaway interface {
	Reconnect(sessionId uint64, handleId uint64, plugin string) (IHandle, ISession, error)
	Info() (*InfoMsg, error)
	Create() (ISession, error)
	Close() error
}

// Gateway represents a connection to an instance of the Janus Gateway.
type Gateway struct {
	// Sessions is a map of the currently active sessions to the gateway.
	Sessions map[uint64]*Session

	// Access to the Sessions map should be synchronized with the Gateway.Lock()
	// and Gateway.Unlock() methods provided by the embeded sync.Mutex.
	sync.Mutex

	conn            *websocket.Conn
	httpConn        *http.Client
	httpPath        string
	nextTransaction uint64
	transactions    map[uint64]chan interface{}

	sendChan chan []byte
	writeMu  sync.Mutex
}

//ConnectHttp ...
func ConnectHttp(url string) (*Gateway, error) {
	gateway := new(Gateway)
	gateway.httpConn = &http.Client{
		Timeout: httpTimeout,
	}
	gateway.httpPath = url
	gateway.transactions = make(map[uint64]chan interface{})
	gateway.Sessions = make(map[uint64]*Session)

	gateway.sendChan = make(chan []byte, 100)
	return gateway, nil
}

func Connect(wsURL string) (*Gateway, error) {
	websocket.DefaultDialer.Subprotocols = []string{"janus-protocol"}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	if err != nil {
		return nil, err
	}

	gateway := new(Gateway)
	gateway.conn = conn
	gateway.transactions = make(map[uint64]chan interface{})
	gateway.Sessions = make(map[uint64]*Session)

	gateway.sendChan = make(chan []byte, 100)

	go gateway.ping()
	go gateway.recv()
	return gateway, nil
}

func newSession(id uint64, gw *Gateway) *Session {
	session := new(Session)
	session.id = id
	session.events = make(chan interface{}, 20)
	session.gateway = gw
	session.Handles = make(map[uint64]*Handle)
	session.CloseCh = make(chan bool)
	session.CloseChLock = &sync.Mutex{}
	return session
}

//Reconnect ...
func (gateway *Gateway) Reconnect(sessionId uint64, handleId uint64, plugin string) (IHandle, ISession, error) {
	req, ch := newRequest("claim")
	req["session_id"] = sessionId
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *SuccessMsg:
	case *ErrorMsg:
		return nil, nil, msg
	}

	// Create new session
	session := newSession(sessionId, gateway)

	// Store this session
	gateway.Lock()
	gateway.Sessions[sessionId] = session
	gateway.Unlock()

	if handleId != 0 {
		h := newHandle(handleId, session)

		session.Lock()
		session.Handles[h.id] = h
		session.Unlock()
		return h, session, nil
	}
	h, err := session.Attach(plugin)
	return h, session, err

}

// Close closes the underlying connection to the Gateway.
func (gateway *Gateway) Close() error {
	if gateway.conn == nil {
		return nil
	}
	return gateway.conn.Close()
}

func sendWs(conn *websocket.Conn, msg []byte) error {
	return conn.WriteMessage(websocket.TextMessage, msg)
}

func sendHttp(client *http.Client, path string, msg map[string]interface{}, data []byte) ([]byte, error) {
	var req *http.Request
	var sessionId, handleId string
	var err error
	if sessId, ok := msg["session_id"]; ok && sessId != "" {
		sessionId, _ = sessId.(string)

	}
	if hId, ok := msg["handle_id"]; ok && hId != "" {
		handleId, _ = hId.(string)
	}

	//make sure we are pointing to the correct endpoint
	if sessionId != "" {
		path = path + "/" + sessionId
		if handleId != "" {
			path = path + "/" + handleId
		}
	}

	if msgType, ok := msg["janus"]; ok && msgType != "info" || msgType != "ping" {
		req, err = createReq(path, data)
	} else {
		req, err = createReq(path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("Error creating request to sfu: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request to sfu: %v", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading sfu response: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error unexpected sfu response code %d: %v", resp.StatusCode, err)
	}
	return respBody, nil
}

func (gateway *Gateway) send(msg map[string]interface{}, transaction chan interface{}) {
	id := atomic.AddUint64(&gateway.nextTransaction, 1)

	msg["transaction"] = strconv.FormatUint(id, 10)
	gateway.Lock()
	gateway.transactions[id] = transaction
	gateway.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("json.Marshal: %s\n", err)
		return
	}

	if debug {
		// log message being sent
		var log bytes.Buffer
		json.Indent(&log, data, ">", "   ")
		log.Write([]byte("\n"))
		log.WriteTo(os.Stdout)
	}

	var resp []byte
	gateway.writeMu.Lock()
	if gateway.conn != nil {
		err = sendWs(gateway.conn, data)
	} else {
		resp, err = sendHttp(gateway.httpConn, gateway.httpPath, msg, data)
		//if using the http client we want to send to the recv channel here when we have the response
		if err == nil {
			err = gateway.recvHttpResp(resp)
		}
		if err != nil {
			//pass through the error message to the transaction that it failed on
			passMsg(transaction, &ErrorMsg{ErrorData{Code: 0, Reason: err.Error()}})
		}
	}
	gateway.writeMu.Unlock()
	if err != nil {
		fmt.Printf("conn.Write: %s\n", err)
		return
	}
}

func passMsg(ch chan interface{}, msg interface{}) {
	ch <- msg
}

func (gateway *Gateway) ping() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := gateway.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(20*time.Second))
			if err != nil {
				log.Println("ping:", err)
				return
			}
		}
	}
}

func (gateway *Gateway) sendloop() {

}

func (gateway *Gateway) recvHttpResp(resp []byte) error {
	// Decode to Msg struct
	var base BaseMsg
	if err := json.Unmarshal(resp, &base); err != nil {
		fmt.Printf("json.Unmarshal: %s\n", err)
		return err
	}

	if debug {
		// log message being sent
		var log bytes.Buffer
		json.Indent(&log, resp, "<", "   ")
		log.Write([]byte("\n"))
		log.WriteTo(os.Stdout)
	}

	typeFunc, ok := msgtypes[base.Type]
	if !ok {
		fmt.Printf("Unknown message type received!\n")
		return errors.New("Unknown base type received")
	}

	msg := typeFunc()
	if err := json.Unmarshal(resp, &msg); err != nil {
		fmt.Printf("json.Unmarshal: %s\n", err)
		return err
	}

	// Pass message on from here
	if base.ID == "" {
		// Is this a Handle event?
		if base.Handle == 0 {
			// Error()
		} else if base.Type == "trickle" || base.Type == "webrtcup" {
			// Lookup Session
			gateway.Lock()
			session := gateway.Sessions[base.Session]
			gateway.Unlock()
			if session == nil {
				fmt.Printf("Unable to deliver message. Session gone?\n")
				return errors.New("Unable to deliver message")
			}
			// Pass msg to session
			go passMsg(session.events, msg)
		} else {
			// Lookup Session
			gateway.Lock()
			session := gateway.Sessions[base.Session]
			gateway.Unlock()
			if session == nil {
				fmt.Printf("Unable to deliver message. Session gone?\n")
				return errors.New("Unable to deliver message")
			}

			// Lookup Handle
			session.Lock()
			handle := session.Handles[base.Handle]
			session.Unlock()
			if handle == nil {
				fmt.Printf("Unable to deliver message. Handle gone?\n")
				return errors.New("Unable to deliver message")
			}

			// Pass msg
			go passMsg(handle.Events, msg)
		}
	} else {
		id, _ := strconv.ParseUint(base.ID, 10, 64)
		// Lookup Transaction
		gateway.Lock()
		transaction := gateway.transactions[id]
		gateway.Unlock()
		if transaction == nil {
			// Error()
		}

		// Pass msg
		go passMsg(transaction, msg)
	}
	return nil
}

func (gateway *Gateway) recv() {

	for {
		// Read message from Gateway

		// Decode to Msg struct
		var base BaseMsg

		_, data, err := gateway.conn.ReadMessage()
		if err != nil {
			fmt.Printf("conn.Read: %s\n", err)
			return
		}

		if err := json.Unmarshal(data, &base); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue
		}

		if debug {
			// log message being sent
			var log bytes.Buffer
			json.Indent(&log, data, "<", "   ")
			log.Write([]byte("\n"))
			log.WriteTo(os.Stdout)
		}

		typeFunc, ok := msgtypes[base.Type]
		if !ok {
			fmt.Printf("Unknown message type received!\n")
			continue
		}

		msg := typeFunc()
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Printf("json.Unmarshal: %s\n", err)
			continue // Decode error
		}

		// Pass message on from here
		if base.ID == "" {
			// Is this a Handle event?
			if base.Handle == 0 {
				// Error()
			} else if base.Type == "trickle" || base.Type == "webrtcup" {
				// Lookup Session
				gateway.Lock()
				session := gateway.Sessions[base.Session]
				gateway.Unlock()
				if session == nil {
					fmt.Printf("Unable to deliver message. Session gone?\n")
					continue
				}
				// Pass msg to session
				go passMsg(session.events, msg)
			} else {
				// Lookup Session
				gateway.Lock()
				session := gateway.Sessions[base.Session]
				gateway.Unlock()
				if session == nil {
					fmt.Printf("Unable to deliver message. Session gone?\n")
					continue
				}

				// Lookup Handle
				session.Lock()
				handle := session.Handles[base.Handle]
				session.Unlock()
				if handle == nil {
					fmt.Printf("Unable to deliver message. Handle gone?\n")
					continue
				}

				// Pass msg
				go passMsg(handle.Events, msg)
			}
		} else {
			id, _ := strconv.ParseUint(base.ID, 10, 64)
			// Lookup Transaction
			gateway.Lock()
			transaction := gateway.transactions[id]
			gateway.Unlock()
			if transaction == nil {
				// Error()
			}

			// Pass msg
			go passMsg(transaction, msg)
		}
	}
}

// Info sends an info request to the Gateway.
// On success, an InfoMsg will be returned and error will be nil.
func (gateway *Gateway) Info() (*InfoMsg, error) {
	req, ch := newRequest("info")
	gateway.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *InfoMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("info")
}

// Create sends a create request to the Gateway.
// On success, a new Session will be returned and error will be nil.
func (gateway *Gateway) Create() (ISession, error) {
	req, ch := newRequest("create")
	gateway.send(req, ch)

	msg := <-ch
	var success *SuccessMsg
	switch msg := msg.(type) {
	case *SuccessMsg:
		success = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Create new session
	session := newSession(success.Data.ID, gateway)

	// Store this session
	gateway.Lock()
	gateway.Sessions[session.id] = session
	gateway.Unlock()

	return session, nil
}

type ISession interface {
	GetId() uint64
	Events() <-chan interface{}
	Attach(plugin string) (IHandle, error)
	Destroy() (*AckMsg, error)
	StopPoll()
	LongPollForEvents()
	KeepAlive() (*AckMsg, error)
}

// Session represents a session instance on the Janus Gateway.
type Session struct {
	// id is the session_id of this session
	id uint64

	// Handles is a map of plugin handles within this session
	Handles map[uint64]*Handle

	events chan interface{}

	// Access to the Handles map should be synchronized with the Session.Lock()
	// and Session.Unlock() methods provided by the embeded sync.Mutex.
	sync.Mutex

	CloseCh     chan bool
	CloseChLock *sync.Mutex
	gateway     *Gateway
}

//LongPollForEvents ...
func (session *Session) LongPollForEvents() {
	msg := map[string]interface{}{
		"session_id": strconv.Itoa(int(session.id)),
		"janus":      "ping",
	}
	var err error
	var resp []byte
	for {
		select {
		case <-session.CloseCh:
			return
		default:
		}
		resp, err = sendHttp(session.gateway.httpConn, session.gateway.httpPath, msg, nil)
		if err == nil {
			session.gateway.recvHttpResp(resp)
		} else if strings.Contains(err.Error(), "context deadline exceeded") {
			return
		}
	}
}

//GetId ...
func (session *Session) GetId() uint64 {
	return session.id
}

//Events ...
func (session *Session) Events() <-chan interface{} {
	return session.events
}

func (session *Session) send(msg map[string]interface{}, transaction chan interface{}) {
	msg["session_id"] = session.id
	session.gateway.send(msg, transaction)
}

func newHandle(id uint64, sess *Session) *Handle {
	handle := new(Handle)
	handle.id = id
	handle.session = sess
	handle.Events = make(chan interface{}, 8)
	return handle
}

// Attach sends an attach request to the Gateway within this session.
// plugin should be the unique string of the plugin to attach to.
// On success, a new Handle will be returned and error will be nil.
func (session *Session) Attach(plugin string) (IHandle, error) {
	req, ch := newRequest("attach")
	req["plugin"] = plugin
	session.send(req, ch)

	var success *SuccessMsg
	msg := <-ch
	switch msg := msg.(type) {
	case *SuccessMsg:
		success = msg
	case *ErrorMsg:
		return nil, msg
	}

	handle := newHandle(success.Data.ID, session)

	session.Lock()
	session.Handles[handle.id] = handle
	session.Unlock()

	return handle, nil
}

// KeepAlive sends a keep-alive request to the Gateway.
// On success, an AckMsg will be returned and error will be nil.
func (session *Session) KeepAlive() (*AckMsg, error) {
	req, ch := newRequest("keepalive")
	session.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("keepalive")
}

// Destroy sends a destroy request to the Gateway to tear down this session.
// On success, the Session will be removed from the Gateway.Sessions map, an
// AckMsg will be returned and error will be nil.
func (session *Session) Destroy() (*AckMsg, error) {
	req, ch := newRequest("destroy")
	session.send(req, ch)

	var ack *AckMsg
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		ack = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Remove this session from the gateway
	session.gateway.Lock()
	delete(session.gateway.Sessions, session.id)
	session.gateway.Unlock()

	return ack, nil
}

//StopPoll ...
func (session *Session) StopPoll() {
	session.CloseChLock.Lock()
	defer session.CloseChLock.Unlock()
	select {
	case <-session.CloseCh:
	default:
		close(session.CloseCh)
	}
}

//IHandle ...
type IHandle interface {
	GetId() uint64
	Request(body interface{}) (*SuccessMsg, error)
	Message(body, jsep interface{}) (*EventMsg, error)
	Trickle(candidate interface{}) (*AckMsg, error)
	TrickleMany(candidates interface{}) (*AckMsg, error)
	Detach() (*AckMsg, error)
}

// Handle represents a handle to a plugin instance on the Gateway.
type Handle struct {
	// id is the handle_id of this plugin handle
	id uint64

	// Type   // pub  or sub
	Type string

	//User   // Userid
	User string

	// Events is a receive only channel that can be used to receive events
	// related to this handle from the gateway.
	Events chan interface{}

	session *Session
}

//GetId ...
func (handle *Handle) GetId() uint64 {
	return handle.id
}

func (handle *Handle) send(msg map[string]interface{}, transaction chan interface{}) {
	msg["handle_id"] = handle.id
	handle.session.send(msg, transaction)
}

// send sync request
func (handle *Handle) Request(body interface{}) (*SuccessMsg, error) {
	req, ch := newRequest("message")
	if body != nil {
		req["body"] = body
	}
	handle.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *SuccessMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("message")
}

// Message sends a message request to a plugin handle on the Gateway.
// body should be the plugin data to be passed to the plugin, and jsep should
// contain an optional SDP offer/answer to establish a WebRTC PeerConnection.
// On success, an EventMsg will be returned and error will be nil.
func (handle *Handle) Message(body, jsep interface{}) (*EventMsg, error) {
	req, ch := newRequest("message")
	if body != nil {
		req["body"] = body
	}
	if jsep != nil {
		req["jsep"] = jsep
	}
	handle.send(req, ch)

GetMessage: // No tears..
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		goto GetMessage // ..only dreams.
	case *EventMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("message")
}

// Trickle sends a trickle request to the Gateway as part of establishing
// a new PeerConnection with a plugin.
// candidate should be a single ICE candidate, or a completed object to
// signify that all candidates have been sent:
//		{
//			"completed": true
//		}
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) Trickle(candidate interface{}) (*AckMsg, error) {
	req, ch := newRequest("trickle")
	req["candidate"] = candidate
	handle.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("trickle")
}

// TrickleMany sends a trickle request to the Gateway as part of establishing
// a new PeerConnection with a plugin.
// candidates should be an array of ICE candidates.
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) TrickleMany(candidates interface{}) (*AckMsg, error) {
	req, ch := newRequest("trickle")
	req["candidates"] = candidates
	handle.send(req, ch)

	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		return msg, nil
	case *ErrorMsg:
		return nil, msg
	}

	return nil, unexpected("trickle")
}

// Detach sends a detach request to the Gateway to remove this handle.
// On success, an AckMsg will be returned and error will be nil.
func (handle *Handle) Detach() (*AckMsg, error) {
	req, ch := newRequest("detach")
	handle.send(req, ch)

	var ack *AckMsg
	msg := <-ch
	switch msg := msg.(type) {
	case *AckMsg:
		ack = msg
	case *ErrorMsg:
		return nil, msg
	}

	// Remove this handle from the session
	handle.session.Lock()
	delete(handle.session.Handles, handle.id)
	handle.session.Unlock()

	return ack, nil
}

func createReq(path string, body []byte) (*http.Request, error) {
	if body != nil {
		return http.NewRequest("POST", path, bytes.NewBuffer(body))
	}

	return http.NewRequest("GET", path, nil)
}
