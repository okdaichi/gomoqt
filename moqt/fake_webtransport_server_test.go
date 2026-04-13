package moqt

var _ WebTransportServer = (*FakeWebTransportServer)(nil)

// FakeWebTransportServer is a fake implementation of the WebTransportServer interface.
type FakeWebTransportServer struct {
	ServeQUICConnFunc func(conn StreamConn) error
	CloseFunc         func() error
}

func (m *FakeWebTransportServer) ServeQUICConn(conn StreamConn) error {
	if m.ServeQUICConnFunc != nil {
		return m.ServeQUICConnFunc(conn)
	}
	return nil
}

func (m *FakeWebTransportServer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
