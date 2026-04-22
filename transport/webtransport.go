package transport

// WebTransportSession represents a WebTransport session and its stream-oriented
// transport connection. It intentionally does not expose connection-level stats
// directly; transports that support stats may offer them through the optional
// ConnectionStatsProvider interface.
type WebTransportSession interface {
	StreamConn
	Subprotocol() string
}
