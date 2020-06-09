package janus

import "sync"

//NewTestSession used for testing
func NewTestSession(id uint64, gw *Gateway) *Session {
	session := new(Session)
	session.id = id
	session.events = make(chan interface{})
	session.gateway = gw
	session.Handles = make(map[uint64]*Handle)
	session.CloseCh = make(chan bool)
	session.CloseChLock = &sync.Mutex{}
	return session
}

//NewTestHandle used for testing
func NewTestHandle(id uint64, sess *Session) *Handle {
	handle := new(Handle)
	handle.id = id
	handle.session = sess
	handle.Events = make(chan interface{}, 8)
	return handle
}
