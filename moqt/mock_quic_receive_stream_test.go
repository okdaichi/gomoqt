package moqt

import (
	"time"

	"github.com/okdaichi/gomoqt/transport"
	"github.com/stretchr/testify/mock"
)

var _ transport.ReceiveStream = (*MockQUICReceiveStream)(nil)

// MockQUICReceiveStream is a mock implementation of ReceiveStream using testify/mock
type MockQUICReceiveStream struct {
	mock.Mock
	ReadFunc            func(p []byte) (n int, err error)
	StreamIDFunc        func() transport.StreamID
	SetReadDeadlineFunc func(t time.Time) error
}

func (m *MockQUICReceiveStream) StreamID() transport.StreamID {
	if m.StreamIDFunc != nil {
		return m.StreamIDFunc()
	}
	for _, call := range m.ExpectedCalls {
		if call.Method != "StreamID" {
			continue
		}
		args := m.Called()
		if len(args) == 0 || args.Get(0) == nil {
			return transport.StreamID(0)
		}
		return args.Get(0).(transport.StreamID)
	}
	return transport.StreamID(0)
}

func (m *MockQUICReceiveStream) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICReceiveStream) CancelRead(code transport.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICReceiveStream) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	for _, call := range m.ExpectedCalls {
		if call.Method != "SetReadDeadline" {
			continue
		}
		args := m.Called(t)
		return args.Error(0)
	}
	return nil
}
