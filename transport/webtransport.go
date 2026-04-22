package transport

// WebTransportSession represents a WebTransport session and its stream-oriented
// transport connection. It intentionally does not expose connection-level stats
// directly; transports that support stats may offer them through an optional
// stats-provider interface implemented by the concrete session type.
type WebTransportSession interface {
	StreamConn
	Subprotocol() string
}
