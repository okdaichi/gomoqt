import type { ReceiveStream, SendStream } from "./internal/webtransport/mod.ts";
import { withCancelCause } from "@okdaichi/golikejs/context";
import type { CancelCauseFunc, Context } from "@okdaichi/golikejs/context";
import { WebTransportStreamError } from "./internal/webtransport/error.ts";
import type { GroupMessage } from "./internal/message/mod.ts";
import { readFull, readVarint, writeVarint } from "./internal/message/mod.ts";
import { GroupErrorCode } from "./error.ts";
import { GroupSequence } from "./alias.ts";
import { ByteSink, ByteSinkFunc, ByteSource } from "./frame.ts";

export class GroupWriter {
	readonly sequence: GroupSequence;
	#stream: SendStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(trackCtx: Context, writer: SendStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#stream = writer;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(GroupErrorCode.SubscribeCanceled);
		});
	}

	async writeFrame(src: ByteSource): Promise<Error | undefined> {
		// Convert source to bytes using copyTo
		const buf = new Uint8Array(src.byteLength);
		src.copyTo(buf);
		const bytes = buf;

		// Write length prefix as varint
		let [, err] = await writeVarint(this.#stream, bytes.byteLength);
		if (err) {
			return err;
		}

		// Write frame data
		[, err] = await this.#stream.write(bytes);
		if (err) {
			return err;
		}

		return undefined;
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			return;
		}
		this.#cancelFunc(undefined); // Notify the context about closure
		await this.#stream.close();
	}

	async cancel(code: GroupErrorCode): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const cause = new WebTransportStreamError(
			{ source: "stream", streamErrorCode: code },
			false,
		);
		this.#cancelFunc(cause); // Notify the context about cancellation
		await this.#stream.cancel(code);
	}
}

export class GroupReader {
	readonly sequence: GroupSequence;
	#reader: ReceiveStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(trackCtx: Context, reader: ReceiveStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#reader = reader;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(GroupErrorCode.PublishAborted);
		});
	}

	async readFrame(sink: ByteSink | ByteSinkFunc): Promise<Error | undefined> {
		// Read length prefix as varint
		const [len, , err1] = await readVarint(this.#reader);
		if (err1) {
			return err1;
		}

		try {
			// Create a buffer for reading
			const buf = new Uint8Array(len);

			// Read the frame data
			const [, err2] = await readFull(this.#reader, buf);
			if (err2) {
				return err2;
			}

			// Write to sink (handle both ByteSink and ByteSinkFunc)
			if (typeof sink === "function") {
				await sink(buf);
			} else {
				await sink.write(buf);
			}
		} catch (e) {
			return e instanceof Error ? e : new Error(String(e));
		}

		return undefined;
	}

	async cancel(code: GroupErrorCode): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const reason = new WebTransportStreamError(
			{ source: "stream", streamErrorCode: code },
			false,
		);
		this.#cancelFunc(reason);
		await this.#reader.cancel(code);
	}
}
