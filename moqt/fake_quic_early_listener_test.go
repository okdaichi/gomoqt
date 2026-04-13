package moqt

import (
	"context"
	"net"
)

var _ QUICListener = (*FakeEarlyListener)(nil)

// FakeEarlyListener is a fake implementation of QUICListener.
type FakeEarlyListener struct {
	AcceptFunc func(ctx context.Context) (StreamConn, error)
	CloseFunc  func() error
	AddrFunc   func() net.Addr

	closeCalled bool
}

func (m *FakeEarlyListener) Accept(ctx context.Context) (StreamConn, error) {
	if m.AcceptFunc != nil {
		return m.AcceptFunc(ctx)
	}
	return nil, nil
}

func (m *FakeEarlyListener) Addr() net.Addr {
	if m.AddrFunc != nil {
		return m.AddrFunc()
	}
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *FakeEarlyListener) Close() error {
	m.closeCalled = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
