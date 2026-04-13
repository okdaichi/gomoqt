package moqt

import (
	"io"
	"sync"
	"time"

	"github.com/okdaichi/gomoqt/transport"
)

var _ transport.ReceiveStream = (*FakeQUICReceiveStream)(nil)

// FakeQUICReceiveStream is a fake implementation of ReceiveStream for testing.
// Models quic-go behavior: CancelRead makes subsequent Read return *transport.StreamError.
type FakeQUICReceiveStream struct {
	mu sync.Mutex

	ReadFunc            func(p []byte) (int, error)
	CancelReadFunc      func(transport.StreamErrorCode)
	SetReadDeadlineFunc func(time.Time) error

	cancelReadErr error
}

func (m *FakeQUICReceiveStream) Read(p []byte) (int, error) {
	m.mu.Lock()
	if m.cancelReadErr != nil {
		err := m.cancelReadErr
		m.mu.Unlock()
		return 0, err
	}
	readFunc := m.ReadFunc
	m.mu.Unlock()
	if readFunc != nil {
		return readFunc(p)
	}
	return 0, io.EOF
}

func (m *FakeQUICReceiveStream) CancelRead(code transport.StreamErrorCode) {
	if m.CancelReadFunc != nil {
		m.CancelReadFunc(code)
		return
	}
	m.mu.Lock()
	if m.cancelReadErr == nil {
		m.cancelReadErr = &transport.StreamError{ErrorCode: code}
	}
	m.mu.Unlock()
}

func (m *FakeQUICReceiveStream) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	return nil
}
