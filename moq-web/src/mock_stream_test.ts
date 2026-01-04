/**
 * Mock implementations for Stream, SendStream, and ReceiveStream interfaces.
 * These mocks are type-safe and avoid the need for `as any` casts in tests.
 */

import { spy } from "@std/testing/mock";
import type { Stream } from "./internal/webtransport/stream.ts";
import type { SendStream } from "./internal/webtransport/send_stream.ts";
import type { ReceiveStream } from "./internal/webtransport/receive_stream.ts";
import { EOFError } from "@okdaichi/golikejs/io";

/**
 * Mock SendStream that implements the SendStream interface.
 * Accepts Partial<SendStream> to override default implementations.
 */
export class MockSendStream implements SendStream {
	readonly id: bigint;
	readonly write: (p: Uint8Array) => Promise<[number, Error | undefined]>;
	readonly close: () => Promise<void>;
	readonly cancel: (code: number) => Promise<void>;
	readonly closed: () => Promise<void>;

	constructor(partial: Partial<SendStream> = {}) {
		this.id = partial.id ?? 0n;
		const writeFunc = partial.write ?? (async (p: Uint8Array) => [p.length, undefined] as [number, Error | undefined]);
		this.write = spy(async (p: Uint8Array) => await writeFunc(p));
		const closeFunc = partial.close ?? (async () => {});
		this.close = spy(async () => await closeFunc());
		const cancelFunc = partial.cancel ?? (async () => {});
		this.cancel = spy(async (code: number) => await cancelFunc(code));
		this.closed = partial.closed ?? (() => new Promise<void>(() => {}));
	}
}

/**
 * Mock ReceiveStream that implements the ReceiveStream interface.
 * Accepts Partial<ReceiveStream> to override default implementations.
 */
export class MockReceiveStream implements ReceiveStream {
	readonly id: bigint;
	readonly read: (p: Uint8Array) => Promise<[number, Error | undefined]>;
	readonly cancel: (code: number) => Promise<void>;
	readonly closed: () => Promise<void>;

	constructor(partial: Partial<ReceiveStream> = {}) {
		this.id = partial.id ?? 0n;
		const readFunc = partial.read ?? (async () => [0, new EOFError()] as [number, Error | undefined]);
		this.read = spy(async (p: Uint8Array) => await readFunc(p));
		const cancelFunc = partial.cancel ?? (async () => {});
		this.cancel = spy(async (code: number) => await cancelFunc(code));
		this.closed = partial.closed ?? (() => new Promise<void>(() => {}));
	}
}

/**
 * Mock Stream that implements the Stream interface.
 * Accepts Partial<Stream> to override default implementations.
 */
export class MockStream implements Stream {
	readonly id: bigint;
	readonly writable: SendStream;
	readonly readable: ReceiveStream;

	constructor(partial: Partial<Stream> = {}) {
		this.id = partial.id ?? 0n;
		this.writable = partial.writable ?? new MockSendStream({ id: partial.id });
		this.readable = partial.readable ??
			new MockReceiveStream({ id: partial.id });
	}
}
