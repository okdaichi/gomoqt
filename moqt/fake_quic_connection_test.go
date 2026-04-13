package moqt

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/okdaichi/gomoqt/transport"
	quicgo "github.com/quic-go/quic-go"
)

var _ StreamConn = (*MockStreamConn)(nil)

// MockStreamConn is a fake implementation of StreamConn.
// All methods delegate to corresponding Func fields when set,
// otherwise return sensible zero-value defaults.
type MockStreamConn struct {
	AcceptStreamFunc      func(ctx context.Context) (transport.Stream, error)
	AcceptUniStreamFunc   func(ctx context.Context) (transport.ReceiveStream, error)
	OpenStreamFunc        func() (transport.Stream, error)
	OpenUniStreamFunc     func() (transport.SendStream, error)
	OpenStreamSyncFunc    func(ctx context.Context) (transport.Stream, error)
	OpenUniStreamSyncFunc func(ctx context.Context) (transport.SendStream, error)
	CloseWithErrorFunc    func(code transport.ConnErrorCode, reason string) error
	ParentCtx             context.Context
	LocalAddrFunc         func() net.Addr
	RemoteAddrFunc        func() net.Addr
	TLSFunc               func() *tls.ConnectionState
	ConnectionStatsFunc   func() quicgo.ConnectionStats
}

func (m *MockStreamConn) TLS() *tls.ConnectionState {
	if m.TLSFunc != nil {
		return m.TLSFunc()
	}
	return nil
}

func (m *MockStreamConn) AcceptStream(ctx context.Context) (transport.Stream, error) {
	if m.AcceptStreamFunc != nil {
		return m.AcceptStreamFunc(ctx)
	}
	return nil, nil
}

func (m *MockStreamConn) AcceptUniStream(ctx context.Context) (transport.ReceiveStream, error) {
	if m.AcceptUniStreamFunc != nil {
		return m.AcceptUniStreamFunc(ctx)
	}
	return nil, nil
}

func (m *MockStreamConn) OpenStream() (transport.Stream, error) {
	if m.OpenStreamFunc != nil {
		return m.OpenStreamFunc()
	}
	return nil, nil
}

func (m *MockStreamConn) OpenUniStream() (transport.SendStream, error) {
	if m.OpenUniStreamFunc != nil {
		return m.OpenUniStreamFunc()
	}
	return nil, nil
}

func (m *MockStreamConn) OpenStreamSync(ctx context.Context) (transport.Stream, error) {
	if m.OpenStreamSyncFunc != nil {
		return m.OpenStreamSyncFunc(ctx)
	}
	return nil, nil
}

func (m *MockStreamConn) OpenUniStreamSync(ctx context.Context) (transport.SendStream, error) {
	if m.OpenUniStreamSyncFunc != nil {
		return m.OpenUniStreamSyncFunc(ctx)
	}
	return nil, nil
}

func (m *MockStreamConn) LocalAddr() net.Addr {
	if m.LocalAddrFunc != nil {
		return m.LocalAddrFunc()
	}
	return &net.TCPAddr{}
}

func (m *MockStreamConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFunc != nil {
		return m.RemoteAddrFunc()
	}
	return &net.TCPAddr{}
}

func (m *MockStreamConn) CloseWithError(code transport.ConnErrorCode, reason string) error {
	if m.CloseWithErrorFunc != nil {
		return m.CloseWithErrorFunc(code, reason)
	}
	return nil
}

func (m *MockStreamConn) Context() context.Context {
	if m.ParentCtx != nil {
		return m.ParentCtx
	}
	return context.Background()
}

func (m *MockStreamConn) ConnectionStats() quicgo.ConnectionStats {
	if m.ConnectionStatsFunc != nil {
		return m.ConnectionStatsFunc()
	}
	return quicgo.ConnectionStats{}
}

type FakeWebTransportSession struct {
	MockStreamConn
	SubprotocolFunc func() string
}

func (m *FakeWebTransportSession) Subprotocol() string {
	if m.SubprotocolFunc != nil {
		return m.SubprotocolFunc()
	}
	return ""
}
