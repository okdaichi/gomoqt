package webtransportgo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// settingsEnableWebtransport is the HTTP/3 settings identifier for WebTransport,
// as defined in github.com/quic-go/webtransport-go/protocol.go.
const settingsEnableWebtransport uint64 = 0x2b603742

// TestNewServer_ConnContextIsSet is a regression test for the bug introduced
// when webtransport-go changed Server.H3 from a value type to *http3.Server.
//
// webtransport-go v0.10.0 requires ConfigureHTTP3Server to be called on H3 so
// that H3.ConnContext injects the *quic.Conn into every HTTP/3 request context.
// Without it, Server.Upgrade() cannot find the connection and returns:
//   "webtransport: missing QUIC connection"
func TestNewServer_ConnContextIsSet(t *testing.T) {
	srv := NewServer(nil)

	wrapper, ok := srv.(*serverWrapper)
	require.True(t, ok, "NewServer must return a *serverWrapper")

	assert.NotNil(t, wrapper.server.H3.ConnContext,
		"H3.ConnContext must be non-nil after ConfigureHTTP3Server; "+
			"a nil ConnContext causes every Upgrade() call to fail with "+
			"\"webtransport: missing QUIC connection\"")
}

// TestNewServer_EnableDatagramsIsSet verifies that H3.EnableDatagrams is true,
// which is required for HTTP/3-level QUIC datagram support used by WebTransport.
func TestNewServer_EnableDatagramsIsSet(t *testing.T) {
	srv := NewServer(nil)

	wrapper, ok := srv.(*serverWrapper)
	require.True(t, ok)

	assert.True(t, wrapper.server.H3.EnableDatagrams,
		"H3.EnableDatagrams must be true for WebTransport")
}

// TestNewServer_WebTransportSettingAdvertised verifies that the HTTP/3 SETTINGS
// frame will advertise WebTransport support to clients.
func TestNewServer_WebTransportSettingAdvertised(t *testing.T) {
	srv := NewServer(nil)

	wrapper, ok := srv.(*serverWrapper)
	require.True(t, ok)

	require.NotNil(t, wrapper.server.H3.AdditionalSettings,
		"H3.AdditionalSettings must not be nil")

	val, exists := wrapper.server.H3.AdditionalSettings[settingsEnableWebtransport]
	assert.True(t, exists,
		"H3.AdditionalSettings must contain settingsEnableWebtransport (0x2b603742)")
	assert.Equal(t, uint64(1), val,
		"settingsEnableWebtransport must be set to 1")
}

// TestNewServer_CheckOriginPropagated verifies that the checkOrigin function
// provided by the caller is forwarded to the underlying webtransport-go Server.
func TestNewServer_CheckOriginPropagated(t *testing.T) {
	called := false
	checkOrigin := func(r *http.Request) bool {
		called = true
		return true
	}

	srv := NewServer(checkOrigin)

	wrapper, ok := srv.(*serverWrapper)
	require.True(t, ok)

	require.NotNil(t, wrapper.server.CheckOrigin,
		"CheckOrigin must be propagated to the underlying server")

	// Invoke the stored function to confirm it's the one we passed in.
	wrapper.server.CheckOrigin(&http.Request{})
	assert.True(t, called, "stored CheckOrigin should be our function")
}

// TestNewServer_NilCheckOriginDoesNotPanic verifies that passing nil for
// checkOrigin is safe (the underlying library substitutes its own default).
func TestNewServer_NilCheckOriginDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		NewServer(nil)
	})
}
