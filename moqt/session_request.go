package moqt

import (
	\ github.com/okdaichi/gomoqt/quic\
)

// SessionRequest represents an incoming connection before a MOQ session is established.
type SessionRequest struct {
	conn quic.Connection
	srv  *Server
}

// Connection returns the underlying QUIC connection.
func (r *SessionRequest) Connection() quic.Connection {
	return r.conn
}

// Accept accepts the incoming connection and establishes a MOQ session with the given TrackMux.
func Accept(req *SessionRequest, mux *TrackMux) (*Session, error) {
	if mux == nil {
		mux = DefaultMux
	}

	var sess *Session
	sess = newSession(
		req.conn,
		mux,
		func() {
			if req.srv != nil {
				req.srv.removeSession(sess)
			}
		},
	)

	if req.srv != nil {
		req.srv.addSession(sess)
	}

	return sess, nil
}

