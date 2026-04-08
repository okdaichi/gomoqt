package moqt

import "sync"

type groupManager struct {
	mu           sync.Mutex
	activeGroups map[*GroupWriter]struct{}

	closed bool
}

func newGroupManager() *groupManager {
	return &groupManager{
		activeGroups: make(map[*GroupWriter]struct{}),
	}
}

func (m *groupManager) addGroup(group *GroupWriter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.activeGroups[group] = struct{}{}
}

func (m *groupManager) removeGroup(group *GroupWriter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeGroups, group)
}

func (m *groupManager) countGroups() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.activeGroups)
}

func (m *groupManager) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.activeGroups = nil
}
