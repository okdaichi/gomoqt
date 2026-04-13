package webtransportgo

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/okdaichi/gomoqt/transport"
)

var _ transport.StreamConn = (*FakeStreamConn)(nil)

type FakeStreamConn struct {
	AcceptStreamFunc      func(ctx context.Context) (transport.Stream, error)
	AcceptUniStreamFunc   func(ctx context.Context) (transport.ReceiveStream, error)
	CloseWithErrorFunc    func(code transport.ConnErrorCode, msg string) error
	ParentCtx             context.Context
	LocalAddrFunc         func() net.Addr
	RemoteAddrFunc        func() net.Addr
	OpenStreamFunc        func() (transport.Stream, error)
	OpenStreamSyncFunc    func(ctx context.Context) (transport.Stream, error)
	OpenUniStreamFunc     func() (transport.SendStream, error)
	OpenUniStreamSyncFunc func(ctx context.Context) (transport.SendStream, error)
	TLSFunc               func() *tls.ConnectionState
}

func (m *FakeStreamConn) AcceptStream(ctx context.Context) (transport.Stream, error) {
	if m.AcceptStreamFunc != nil {
		return m.AcceptStreamFunc(ctx)
	}
	return nil, nil
}

func (m *FakeStreamConn) AcceptUniStream(ctx context.Context) (transport.ReceiveStream, error) {
	if m.AcceptUniStreamFunc != nil {
		return m.AcceptUniStreamFunc(ctx)
	}
	return nil, nil
}

func (m *FakeStreamConn) CloseWithError(code transport.ConnErrorCode, msg string) error {
	if m.CloseWithErrorFunc != nil {
		return m.CloseWithErrorFunc(code, msg)
	}
	return nil
}

func (m *FakeStreamConn) Context() context.Context {
	if m.ParentCtx != nil {
		return m.ParentCtx
	}
	return context.Background()
}

func (m *FakeStreamConn) LocalAddr() net.Addr {
	if m.LocalAddrFunc != nil {
		return m.LocalAddrFunc()
	}
	return &net.TCPAddr{}
}

func (m *FakeStreamConn) OpenStream() (transport.Stream, error) {
	if m.OpenStreamFunc != nil {
		return m.OpenStreamFunc()
	}
	return nil, nil
}

func (m *FakeStreamConn) OpenStreamSync(ctx context.Context) (transport.Stream, error) {
	if m.OpenStreamSyncFunc != nil {
		return m.OpenStreamSyncFunc(ctx)
	}
	return nil, nil
}

func (m *FakeStreamConn) OpenUniStream() (transport.SendStream, error) {
	if m.OpenUniStreamFunc != nil {
		return m.OpenUniStreamFunc()
	}
	return nil, nil
}

func (m *FakeStreamConn) OpenUniStreamSync(ctx context.Context) (transport.SendStream, error) {
	if m.OpenUniStreamSyncFunc != nil {
		return m.OpenUniStreamSyncFunc(ctx)
	}
	return nil, nil
}

func (m *FakeStreamConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFunc != nil {
		return m.RemoteAddrFunc()
	}
	return &net.TCPAddr{}
}

func (m *FakeStreamConn) TLS() *tls.ConnectionState {
	if m.TLSFunc != nil {
		return m.TLSFunc()
	}
	return nil
}
