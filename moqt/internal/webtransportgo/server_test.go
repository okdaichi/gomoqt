package webtransportgo

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/okdaichi/gomoqt/transport"
	"github.com/stretchr/testify/require"
)

// TestServer_Init_CreatesInternalServer verifies wrapper init allocates the
// upstream server and an HTTP/3 server container.
func TestServer_Init_CreatesInternalServer(t *testing.T) {
	srv := &Server{}
	srv.init()

	require.NotNil(t, srv.internalServer)
	require.NotNil(t, srv.internalServer.H3)
}

// TestInit_DoesNotPanic verifies that init remains safe with default configuration.
func TestServer_Init_DoesNotPanic(t *testing.T) {
	srv := &Server{}
	require.NotPanics(t, func() {
		srv.init()
	})
}

func TestServer_Init_SetsConnContextWhenProvided(t *testing.T) {
	type testKey struct{}

	srv := &Server{
		ConnContext: func(ctx context.Context, conn transport.StreamConn) context.Context {
			require.Nil(t, conn)
			return context.WithValue(ctx, testKey{}, "ok")
		},
	}
	srv.init()

	require.NotNil(t, srv.internalServer)
	require.NotNil(t, srv.internalServer.H3)
	require.NotNil(t, srv.internalServer.H3.ConnContext)

	ctx := srv.internalServer.H3.ConnContext(context.Background(), nil)
	require.Equal(t, "ok", ctx.Value(testKey{}))
}

func TestServer_Init_PanicsOnNilConnContextResult(t *testing.T) {
	srv := &Server{
		ConnContext: func(ctx context.Context, conn transport.StreamConn) context.Context {
			return nil
		},
	}
	srv.init()

	require.Panics(t, func() {
		_ = srv.internalServer.H3.ConnContext(context.Background(), nil)
	})
}

type dummyStreamConn struct{}

func (dummyStreamConn) AcceptStream(context.Context) (transport.Stream, error) { return nil, errors.New("not implemented") }
func (dummyStreamConn) AcceptUniStream(context.Context) (transport.ReceiveStream, error) {
	return nil, errors.New("not implemented")
}
func (dummyStreamConn) CloseWithError(code transport.ConnErrorCode, msg string) error { return nil }
func (dummyStreamConn) Context() context.Context                                { return context.Background() }
func (dummyStreamConn) LocalAddr() net.Addr                                     { return &net.TCPAddr{} }
func (dummyStreamConn) OpenStream() (transport.Stream, error)                   { return nil, errors.New("not implemented") }
func (dummyStreamConn) OpenStreamSync(context.Context) (transport.Stream, error) {
	return nil, errors.New("not implemented")
}
func (dummyStreamConn) OpenUniStream() (transport.SendStream, error) { return nil, errors.New("not implemented") }
func (dummyStreamConn) OpenUniStreamSync(context.Context) (transport.SendStream, error) {
	return nil, errors.New("not implemented")
}
func (dummyStreamConn) RemoteAddr() net.Addr  { return &net.TCPAddr{} }
func (dummyStreamConn) TLS() *tls.ConnectionState { return &tls.ConnectionState{} }

func TestServer_ServeQUICConn_InvalidConnType(t *testing.T) {
	srv := &Server{}
	err := srv.ServeQUICConn(dummyStreamConn{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid connection type")
}

func TestServer_ServeQUICConn_NilConn(t *testing.T) {
	srv := &Server{}
	require.NoError(t, srv.ServeQUICConn(nil))
}

func TestServer_Shutdown_WithCancelledContext(t *testing.T) {
	srv := &Server{}
	srv.init()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := srv.Shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestServer_Shutdown_Completes(t *testing.T) {
	srv := &Server{}
	srv.init()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	require.NoError(t, srv.Shutdown(ctx))
}
