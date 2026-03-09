package msf

import (
	"github.com/okdaichi/gomoqt/moqt"
	"github.com/stretchr/testify/mock"
)

var _ moqt.TrackHandler = (*MockTrackHandler)(nil)

// MockTrackHandler is a testify-based mock for moqt.TrackHandler.
type MockTrackHandler struct {
	mock.Mock
}

// ServeTrack implements moqt.TrackHandler.
func (m *MockTrackHandler) ServeTrack(tw *moqt.TrackWriter) {
	m.Called(tw)
}
