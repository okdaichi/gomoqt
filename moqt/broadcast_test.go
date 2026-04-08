package moqt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastRegisterAndServeTrack(t *testing.T) {
	broadcast := NewBroadcast()

	received := make(chan *TrackWriter, 1)
	err := broadcast.Register("video", TrackHandlerFunc(func(tw *TrackWriter) {
		received <- tw
	}))
	require.NoError(t, err)

	tw := &TrackWriter{TrackName: "video"}
	broadcast.ServeTrack(tw)

	select {
	case got := <-received:
		assert.Same(t, tw, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected registered handler to receive track writer")
	}
}

func TestBroadcastRegisterRejectsInvalidInput(t *testing.T) {
	tests := map[string]struct {
		name         TrackName
		handler      TrackHandler
		errorMessage string
	}{
		"empty track name": {
			name:         "",
			handler:      TrackHandlerFunc(func(*TrackWriter) {}),
			errorMessage: "track name is required",
		},
		"nil handler": {
			name:         "video",
			handler:      nil,
			errorMessage: "track handler cannot be nil",
		},
		"typed nil handler func": {
			name:         "video",
			handler:      TrackHandlerFunc(nil),
			errorMessage: "track handler function cannot be nil",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			broadcast := NewBroadcast()
			err := broadcast.Register(tt.name, tt.handler)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)
		})
	}
}

func TestBroadcastRemoveClosesActiveTracks(t *testing.T) {
	broadcast := NewBroadcast()

	started := make(chan struct{})
	closed := make(chan struct{})
	done := make(chan struct{})

	err := broadcast.Register("video", TrackHandlerFunc(func(tw *TrackWriter) {
		close(started)
		<-closed
	}))
	require.NoError(t, err)

	tw := &TrackWriter{
		TrackName:    "video",
		groupManager: newGroupManager(),
		onCloseTrackFunc: func() {
			close(closed)
		},
	}

	go func() {
		broadcast.ServeTrack(tw)
		close(done)
	}()

	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected handler to start serving track")
	}

	assert.True(t, broadcast.Remove("video"))

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected active track to stop after handler removal")
	}
}

func TestBroadcastRegisterReplacementClosesPreviousActiveTracks(t *testing.T) {
	broadcast := NewBroadcast()

	started := make(chan struct{})
	closed := make(chan struct{})
	oldDone := make(chan struct{})
	newCalled := make(chan *TrackWriter, 1)

	err := broadcast.Register("video", TrackHandlerFunc(func(tw *TrackWriter) {
		close(started)
		<-closed
	}))
	require.NoError(t, err)

	oldTrackWriter := &TrackWriter{
		TrackName:    "video",
		groupManager: newGroupManager(),
		onCloseTrackFunc: func() {
			close(closed)
		},
	}

	go func() {
		broadcast.ServeTrack(oldTrackWriter)
		close(oldDone)
	}()

	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected original handler to start serving track")
	}

	err = broadcast.Register("video", TrackHandlerFunc(func(tw *TrackWriter) {
		newCalled <- tw
	}))
	require.NoError(t, err)

	select {
	case <-oldDone:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected previous active track to stop after handler replacement")
	}

	newTrackWriter := &TrackWriter{TrackName: "video"}
	broadcast.ServeTrack(newTrackWriter)

	select {
	case got := <-newCalled:
		assert.Same(t, newTrackWriter, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected replacement handler to receive track writer")
	}
}
