/** Options for configuring a {@link Client}. */
export interface MOQOptions {
	/** Whether the client should automatically reconnect on connection loss. */
	reconnect?: boolean;
	/** Low-level WebTransport options forwarded to the `WebTransport` constructor. */
	transportOptions?: WebTransportOptions;
	/** Called when the server requests session migration via GOAWAY. */
	onGoaway?: (newSessionURI: string) => void;
}
