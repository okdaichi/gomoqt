import type { Reader, Writer } from "@okdaichi/golikejs/io";
import {
	parseVarint,
	readFull,
	readVarint,
	varintLen,
	writeVarint,
} from "./message.ts";

export interface SubscribeOkMessageInit {
	publisherPriority?: number;
	publisherOrdered?: number;
	publisherMaxLatency?: number;
	startGroup?: number;
	endGroup?: number;
}

export class SubscribeOkMessage {
	publisherPriority: number;
	publisherOrdered: number;
	publisherMaxLatency: number;
	startGroup: number;
	endGroup: number;

	constructor(init: SubscribeOkMessageInit = {}) {
		this.publisherPriority = init.publisherPriority ?? 0;
		this.publisherOrdered = init.publisherOrdered ?? 0;
		this.publisherMaxLatency = init.publisherMaxLatency ?? 0;
		this.startGroup = init.startGroup ?? 0;
		this.endGroup = init.endGroup ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.publisherPriority) +
			varintLen(this.publisherOrdered) +
			varintLen(this.publisherMaxLatency) +
			varintLen(this.startGroup) +
			varintLen(this.endGroup)
		);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeVarint(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.publisherPriority);
		if (err) return err;

		[, err] = await writeVarint(w, this.publisherOrdered);
		if (err) return err;

		[, err] = await writeVarint(w, this.publisherMaxLatency);
		if (err) return err;

		[, err] = await writeVarint(w, this.startGroup);
		if (err) return err;

		[, err] = await writeVarint(w, this.endGroup);
		if (err) return err;

		return undefined;
	}

	/**
	 * Decodes the message from the reader.
	 */
	async decode(r: Reader): Promise<Error | undefined> {
		let err: Error | undefined;

		let msgLen: number;
		[msgLen, , err] = await readVarint(r);
		if (err) return err;

		const buf = new Uint8Array(msgLen);
		[, err] = await readFull(r, buf);
		if (err) return err;

		let offset = 0;

		[this.publisherPriority, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.publisherOrdered, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.publisherMaxLatency, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.startGroup, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.endGroup, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		return undefined;
	}
}
