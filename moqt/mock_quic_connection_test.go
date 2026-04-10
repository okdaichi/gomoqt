package moqt

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/okdaichi/gomoqt/transport"
	quicgo "github.com/quic-go/quic-go"
	"github.com/stretchr/testify/mock"
)

var _ StreamConn = (*MockStreamConn)(nil)

// MockStreamConn is a mock implementation of StreamConn using testify/mock
type MockStreamConn struct {
	mock.Mock
	AcceptStreamFunc      func(ctx context.Context) (transport.Stream, error)
	AcceptUniStreamFunc   func(ctx context.Context) (transport.ReceiveStream, error)
	OpenStreamFunc        func() (transport.Stream, error)
	OpenUniStreamFunc     func() (transport.SendStream, error)
	OpenStreamSyncFunc    func(ctx context.Context) (transport.Stream, error)
	OpenUniStreamSyncFunc func(ctx context.Context) (transport.SendStream, error)
	ConnectionStatsFunc   func() quicgo.ConnectionStats
}

// TLS implements [StreamConn].
func (m *MockStreamConn) TLS() *tls.ConnectionState {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*tls.ConnectionState)
}

func (m *MockStreamConn) AcceptStream(ctx context.Context) (transport.Stream, error) {
	if m.AcceptStreamFunc != nil {
		return m.AcceptStreamFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.Stream), args.Error(1)
}

func (m *MockStreamConn) AcceptUniStream(ctx context.Context) (transport.ReceiveStream, error) {
	if m.AcceptUniStreamFunc != nil {
		return m.AcceptUniStreamFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.ReceiveStream), args.Error(1)
}

func (m *MockStreamConn) OpenStream() (transport.Stream, error) {
	if m.OpenStreamFunc != nil {
		return m.OpenStreamFunc()
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.Stream), args.Error(1)
}

func (m *MockStreamConn) OpenUniStream() (transport.SendStream, error) {
	if m.OpenUniStreamFunc != nil {
		return m.OpenUniStreamFunc()
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.SendStream), args.Error(1)
}

func (m *MockStreamConn) OpenStreamSync(ctx context.Context) (transport.Stream, error) {
	if m.OpenStreamSyncFunc != nil {
		return m.OpenStreamSyncFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.Stream), args.Error(1)
}

func (m *MockStreamConn) OpenUniStreamSync(ctx context.Context) (transport.SendStream, error) {
	if m.OpenUniStreamSyncFunc != nil {
		return m.OpenUniStreamSyncFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transport.SendStream), args.Error(1)
}

func (m *MockStreamConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockStreamConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockStreamConn) CloseWithError(code transport.ConnErrorCode, reason string) error {
	args := m.Called(code, reason)
	return args.Error(0)
}

func (m *MockStreamConn) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockStreamConn) ConnectionStats() quicgo.ConnectionStats {
	if m.ConnectionStatsFunc != nil {
		return m.ConnectionStatsFunc()
	}
	return quicgo.ConnectionStats{}
}

type MockWebTransportSession struct {
	MockStreamConn
}

func (m *MockWebTransportSession) Subprotocol() string {
	args := m.Called()
	return args.String(0)
}
