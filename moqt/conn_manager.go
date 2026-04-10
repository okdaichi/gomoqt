package moqt

import (
	"fmt"
	"sync"
)

type connManager struct {
	closed      bool
	mu          sync.Mutex
	connections map[StreamConn]struct{}

	doneChan chan struct{}
}

func newConnManager() *connManager {
	return &connManager{
		connections: make(map[StreamConn]struct{}),
	}
}

func (s *connManager) addConn(conn StreamConn) {
	if conn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	if len(s.connections) == 0 {
		s.doneChan = make(chan struct{})
	}
	s.connections[conn] = struct{}{}
}

func (s *connManager) removeConn(conn StreamConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	delete(s.connections, conn)

	if len(s.connections) == 0 {
		if s.doneChan != nil {
			close(s.doneChan)
			s.doneChan = nil
		}
	}
}

func (s *connManager) countSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.connections)
}

func (s *connManager) Done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.doneChan == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.doneChan
}

func (s *connManager) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	if len(s.connections) != 0 {
		return fmt.Errorf("cannot close session manager with active sessions")
	}
	s.closed = true
	return nil
}
