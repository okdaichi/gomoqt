package webtransportgo

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/okdaichi/gomoqt/webtransport"
	quicgo_quicgo "github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	quicgo_webtransportgo "github.com/quic-go/webtransport-go"
)

func NewServer(checkOrigin func(r *http.Request) bool) webtransport.Server {
	wtserver := &quicgo_webtransportgo.Server{
		H3:          &http3.Server{},
		CheckOrigin: checkOrigin,
	}
	// ConfigureHTTP3Server injects the *quic.Conn into every HTTP/3 request
	// context via H3.ConnContext. Without it, Server.Upgrade() cannot retrieve
	// the QUIC connection and returns "webtransport: missing QUIC connection".
	// It also sets H3.AdditionalSettings[settingsEnableWebtransport]=1 and
	// H3.EnableDatagrams=true which are required by the WebTransport spec.
	quicgo_webtransportgo.ConfigureHTTP3Server(wtserver.H3)

	return wrapServer(wtserver)
}

func wrapServer(server *quicgo_webtransportgo.Server) webtransport.Server {
	return &serverWrapper{
		server: server,
	}
}

var _ webtransport.Server = (*serverWrapper)(nil)

// serverWrapper is a wrapper for Server
type serverWrapper struct {
	server *quicgo_webtransportgo.Server
}

func (wrapper *serverWrapper) Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error) {
	wtsess, err := wrapper.server.Upgrade(w, r)
	if err != nil {
		return nil, err
	}

	return wrapSession(wtsess), nil
}

func (w *serverWrapper) ServeQUICConn(conn quic.Connection) error {
	if conn == nil {
		return nil
	}
	if wrapper, ok := conn.(quicgoUnwrapper); ok {
		return w.server.ServeQUICConn(wrapper.Unwrap())
	}
	return errors.New("invalid connection type: expected a wrapped quic-go connection with Unwrap() method")
}

type quicgoUnwrapper interface {
	Unwrap() *quicgo_quicgo.Conn
}

func (w *serverWrapper) Serve(conn net.PacketConn) error {

	return w.server.Serve(conn)
}

func (w *serverWrapper) Close() error {
	return w.server.Close()
}

func (w *serverWrapper) Shutdown(ctx context.Context) error {
	// Implement a proper shutdown logic that passes the context to the server
	closeCh := make(chan struct{})

	// Close the server in a separate goroutine
	go func() {
		_ = w.server.Close() // Ignore close error as server is shutting down
		close(closeCh)
	}()

	// Wait for either the context to be done or the close to complete
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-closeCh:
		return nil
	}
}
