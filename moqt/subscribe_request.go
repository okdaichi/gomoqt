package moqt

import (
	"fmt"
	"time"
)

// SubscribeRequest represents parameters for one subscribe operation.
type SubscribeRequest struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	// Config holds wire-level subscribe parameters.
	// If nil, a zero-value config is used.
	Config *SubscribeConfig

	// Timeout bounds the subscribe setup wait.
	// If zero, Session default timeout is used.
	Timeout time.Duration

	// OnDrop is invoked when the subscription is dropped by the peer.
	OnDrop func(SubscribeDrop)
}

// NewSubscribeRequest returns a subscribe request initialized with the given values.
func NewSubscribeRequest(path BroadcastPath, name TrackName, config *SubscribeConfig) (*SubscribeRequest, error) {
	if !isValidPath(path) {
		return nil, fmt.Errorf("invalid broadcast path: %q", path)
	}

	req := &SubscribeRequest{
		BroadcastPath: path,
		TrackName:     name,
		Config:        config,
	}
	return req.normalized(), nil
}

func (r *SubscribeRequest) normalized() *SubscribeRequest {
	if r == nil {
		return nil
	}

	cfg := r.Config
	if cfg == nil {
		cfg = &SubscribeConfig{}
	}

	return &SubscribeRequest{
		BroadcastPath: r.BroadcastPath,
		TrackName:     r.TrackName,
		Config:        cfg,
		Timeout:       r.Timeout,
		OnDrop:        r.OnDrop,
	}
}
