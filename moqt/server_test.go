package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/okdaichi/gomoqt/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type stubWTServer struct {
	serveErr error
	closed   bool
}

func (s *stubWTServer) ServeQUICConn(conn transport.StreamConn) error { return s.serveErr }
func (s *stubWTServer) Close() error {
	s.closed = true
	return nil
}

func TestServer_Init(t *testing.T) {
	s := &Server{}
	s.init()

	assert.NotNil(t, s.listeners)
	assert.NotNil(t, s.activeSess)
	assert.NotNil(t, s.doneChan)
}

func TestServer_ServeQUICListener_ShuttingDown(t *testing.T) {
	s := &Server{}
	s.inShutdown.Store(true)

	err := s.ServeQUICListener(&MockEarlyListener{})
	assert.Equal(t, ErrServerClosed, err)
}

func TestServer_ServeQUICConn_UnsupportedProtocol(t *testing.T) {
	s := &Server{}
	conn := &MockStreamConn{}
	conn.On("TLS").Return(&tls.ConnectionState{NegotiatedProtocol: "unknown"})

	err := s.ServeQUICConn(conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol")
}

func TestServer_ServeQUICConn_WebTransport(t *testing.T) {
	s := &Server{WebTransportServer: &stubWTServer{}}
	conn := &MockStreamConn{}
	conn.On("TLS").Return(&tls.ConnectionState{NegotiatedProtocol: NextProtoH3})

	err := s.ServeQUICConn(conn)
	assert.NoError(t, err)
}

func TestServer_ServeQUICConn_NativeQUICHandler(t *testing.T) {
	called := false
	s := &Server{
		NativeQUICHandler: &NativeQUICHandler{
			SessionHandler: func(sess *Session) error {
				called = true
				return nil
			},
		},
	}

	conn := &MockStreamConn{}
	conn.On("TLS").Return(&tls.ConnectionState{NegotiatedProtocol: NextProtoMOQ})
	conn.On("Context").Return(context.Background())
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)

	err := s.ServeQUICConn(conn)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestServer_Close_ClosesListenersAndWTServer(t *testing.T) {
	s := &Server{WebTransportServer: &stubWTServer{}}
	s.init()

	ln := &MockEarlyListener{}
	ln.On("Close").Return(nil)
	s.listeners[ln] = struct{}{}

	err := s.Close()
	assert.NoError(t, err)
	assert.True(t, s.shuttingDown())
	ln.AssertCalled(t, "Close")
	assert.True(t, s.WebTransportServer.(*stubWTServer).closed)
}

func TestServer_Shutdown_NoSessions(t *testing.T) {
	s := &Server{}
	s.init()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.Shutdown(ctx)
	assert.NoError(t, err)
	assert.True(t, s.shuttingDown())
}

func TestServer_addRemoveSession_ShutdownCompletes(t *testing.T) {
	s := &Server{}
	s.init()
	s.inShutdown.Store(true)

	conn := &MockStreamConn{}
	conn.On("Context").Return(context.Background())
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)

	sess := newSession(conn, nil, nil)
	t.Cleanup(func() { _ = sess.CloseWithError(NoError, "") })

	s.addSession(sess)
	s.removeSession(sess)

	select {
	case <-s.doneChan:
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("doneChan should be closed when last session is removed during shutdown")
	}
}

func TestUpgrader_Upgrade_RequiresServerInContext(t *testing.T) {
	u := &Upgrader{}
	r := &http.Request{TLS: &tls.ConnectionState{}}
	_, err := u.Upgrade(&MockHTTPResponseWriter{}, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve server")
}

func TestUpgrader_Upgrade_PlainHTTPRejected(t *testing.T) {
	s := &Server{}
	u := &Upgrader{}

	r, _ := http.NewRequest(http.MethodGet, "https://example.com/moq", nil)
	r = r.WithContext(context.WithValue(context.Background(), serverContextKey, s))
	r.TLS = nil
	r.RemoteAddr = "127.0.0.1:443"

	w := &MockHTTPResponseWriter{}
	w.On("Header").Return(make(http.Header)).Maybe()
	w.On("WriteHeader", http.StatusUpgradeRequired).Maybe()
	w.On("Write", mock.Anything).Return(0, nil).Maybe()

	_, err := u.Upgrade(w, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plain HTTP")
}

func TestUpgrader_Upgrade_Success(t *testing.T) {
	s := &Server{}
	s.init()

	u := &Upgrader{
		TrackMux: NewTrackMux(),
		UpgradeFunc: func(w http.ResponseWriter, r *http.Request) (transport.StreamConn, error) {
			conn := &MockStreamConn{}
			conn.On("Context").Return(context.Background())
			conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
			conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 443})
			conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8443})
			return conn, nil
		},
	}

	r, _ := http.NewRequest(http.MethodGet, "https://example.com/moq", nil)
	r = r.WithContext(context.WithValue(context.Background(), serverContextKey, s))
	r.TLS = &tls.ConnectionState{}

	w := &MockHTTPResponseWriter{}
	w.On("Header").Return(make(http.Header)).Maybe()

	sess, err := u.Upgrade(w, r)
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Len(t, s.activeSess, 1)

	_ = sess.CloseWithError(NoError, "")
}

func TestNativeQUICHandler_NoSessionHandler(t *testing.T) {
	h := &NativeQUICHandler{}
	err := h.handleNativeQUIC(&MockStreamConn{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session handler configured")
}

func TestNativeQUICHandler_WithSessionHandler(t *testing.T) {
	called := false
	h := &NativeQUICHandler{
		SessionHandler: func(sess *Session) error {
			called = true
			return errors.New("session handler error")
		},
	}

	conn := &MockStreamConn{}
	conn.On("Context").Return(context.Background())
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)

	err := h.handleNativeQUIC(conn)
	assert.Error(t, err)
	assert.True(t, called)
}
