package moqt

import (
	"sync"
)

type SessionHandler interface {
	ServeSession(sess *Session)
}

type SessionHandlerFunc func(sess *Session)

func (f SessionHandlerFunc) ServeSession(sess *Session) {
	f(sess)
}

type SessionMux struct {
	mu sync.RWMutex
	m  map[string]SessionHandler
}

func NewSessionMux() *SessionMux {
	return &SessionMux{
		m: make(map[string]SessionHandler),
	}
}

func (mux *SessionMux) Handle(pattern string, handler SessionHandler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if handler == nil {
		panic("moqt: nil handler")
	}
	mux.m[pattern] = handler
}

func (mux *SessionMux) HandleFunc(pattern string, handler func(*Session)) {
	mux.Handle(pattern, SessionHandlerFunc(handler))
}

func (mux *SessionMux) ServeSession(sess *Session) {
	mux.mu.RLock()
	handler, ok := mux.m[sess.Path()]
	mux.mu.RUnlock()
	if !ok {
		sess.CloseWithError(404, "not found") // TODO: Use actual error code
		return
	}
	handler.ServeSession(sess)
}

var DefaultSessionMux = NewSessionMux()

func HandleFunc(pattern string, handler func(*Session)) {
	DefaultSessionMux.HandleFunc(pattern, handler)
}
