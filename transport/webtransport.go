package transport

type WebTransportSession interface {
	StreamConn
	ApplicationProtocol() string
}
