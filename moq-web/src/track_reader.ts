import { GroupReader } from "./group_stream.ts";
import type { Info } from "./info.ts";
import type { Context } from "@okdaichi/golikejs/context";
import { ContextCancelledError, watchPromise } from "@okdaichi/golikejs/context";
import type { SendSubscribeStream, SubscribeDrop, TrackConfig } from "./subscribe_stream.ts";
import type { ReceiveStream } from "./internal/webtransport/mod.ts";
import { GroupMessage } from "./internal/message/mod.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import { Queue } from "./internal/queue.ts";
import type { SubscribeID } from "./alias.ts";

export class TrackReader {
	broadcastPath: BroadcastPath;
	trackName: string;
	#subscribeStream: SendSubscribeStream;
	#queue: Queue<[ReceiveStream, GroupMessage]>;
	#onCloseFunc: () => void;

	constructor(
		broadcastPath: BroadcastPath,
		trackName: string,
		subscribeStream: SendSubscribeStream,
		queue: Queue<[ReceiveStream, GroupMessage]>,
		onCloseFunc: () => void,
	) {
		this.broadcastPath = broadcastPath;
		this.trackName = trackName;
		this.#subscribeStream = subscribeStream;
		this.#queue = queue;
		this.#onCloseFunc = onCloseFunc;
	}

	async acceptGroup(
		signal: Promise<void>,
	): Promise<[GroupReader, undefined] | [undefined, Error]> {
		// Check if context is already cancelled
		const err = this.context.err();
		if (err) {
			return [undefined, err];
		}

		while (true) {
			const ctx = watchPromise(this.context, signal);
			const dequeued = await Promise.race([
				this.#queue.dequeue(),
				ctx.done().then(() => {
					return new ContextCancelledError() as Error;
				}),
				this.context.done().then(() => {
					return new Error(
						`track reader context cancelled: ${this.context.err()?.message}`,
					);
				}),
			]);

			if (dequeued instanceof Error) {
				return [undefined, dequeued];
			}
			if (dequeued === undefined) {
				// This is
				throw new Error("dequeue returned undefined");
			}

			const [reader, msg] = dequeued;

			const group = new GroupReader(this.context, reader, msg);

			return [group, undefined];
		}
	}

	async update(config: TrackConfig): Promise<Error | undefined> {
		return this.#subscribeStream.update(config);
	}

	readInfo(): Info {
		return this.#subscribeStream.info;
	}

	async closeWithError(code: number): Promise<void> {
		await this.#subscribeStream.closeWithError(code);
		this.#onCloseFunc();
	}

	async close(): Promise<void> {
		this.#onCloseFunc();
	}

	get subscribeId(): SubscribeID {
		return this.#subscribeStream.subscribeId;
	}

	get trackConfig(): TrackConfig {
		return this.#subscribeStream.config;
	}

	get context(): Context {
		return this.#subscribeStream.context;
	}

	async *drops(signal: Promise<void>): AsyncGenerator<SubscribeDrop> {
		while (true) {
			const [drop, err] = await this.#acceptDrop(signal);
			if (err) {
				return;
			}
			yield drop;
		}
	}

	async #acceptDrop(
		signal: Promise<void>,
	): Promise<[SubscribeDrop, undefined] | [undefined, Error]> {
		while (true) {
			const drops = this.#subscribeStream.pendingDrops();
			if (drops.length > 0) {
				const drop = drops[0]!;
				// Re-append remaining drops
				for (const d of drops.slice(1)) {
					this.#subscribeStream.appendDrop(d);
				}
				return [drop, undefined];
			}

			const ctxErr = this.context.err();
			if (ctxErr) {
				return [undefined, ctxErr];
			}

			const result = await Promise.race([
				signal.then(() => "signal" as const),
				this.context.done().then(() => "ctx" as const),
				this.#subscribeStream.droppedSignal().then(() => "drop" as const),
			]);

			if (result === "signal") {
				return [undefined, new Error("signal cancelled")];
			}

			if (result === "ctx") {
				return [undefined, this.context.err() ?? new Error("context cancelled")];
			}

			// result === "drop" → loop again to dequeue
		}
	}
}
