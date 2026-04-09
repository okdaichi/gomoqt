package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscribeRequest(t *testing.T) {
	req := NewSubscribeRequest("/path", "track", nil)
	require.NotNil(t, req)
	assert.Equal(t, BroadcastPath("/path"), req.BroadcastPath)
	assert.Equal(t, TrackName("track"), req.TrackName)
	require.NotNil(t, req.Config)
	assert.Equal(t, TrackPriority(0), req.Config.Priority)
}

func TestSubscribeRequest_Normalized_DefaultConfig(t *testing.T) {
	req := &SubscribeRequest{
		BroadcastPath: "/path",
		TrackName:     "track",
		Config:        nil,
	}
	norm := req.normalized()
	require.NotNil(t, norm)
	require.NotNil(t, norm.Config)
	assert.Equal(t, TrackPriority(0), norm.Config.Priority)
}
