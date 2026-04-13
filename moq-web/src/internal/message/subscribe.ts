import type { Reader, Writer } from "@okdaichi/golikejs/io";
import {
	parseString,
	parseUint8,
	parseVarint,
	readFull,
	readVarint,
	stringLen,
	varintLen,
	writeString,
	writeUint8,
	writeVarint,
} from "./message.ts";

export interface SubscribeMessageInit {
	subscribeId?: number;
	broadcastPath?: string;
	trackName?: string;
	subscriberPriority?: number;
	subscriberOrdered?: number;
	subscriberMaxLatency?: number;
	startGroup?: number;
	endGroup?: number;
}

export class SubscribeMessage {
	subscribeId: number;
	broadcastPath: string;
	trackName: string;
	subscriberPriority: number;
	subscriberOrdered: number;
	subscriberMaxLatency: number;
	startGroup: number;
	endGroup: number;

	constructor(init: SubscribeMessageInit = {}) {
		this.subscribeId = init.subscribeId ?? 0;
		this.broadcastPath = init.broadcastPath ?? "";
		this.trackName = init.trackName ?? "";
		this.subscriberPriority = init.subscriberPriority ?? 0;
		this.subscriberOrdered = init.subscriberOrdered ?? 0;
		this.subscriberMaxLatency = init.subscriberMaxLatency ?? 0;
		this.startGroup = init.startGroup ?? 0;
		this.endGroup = init.endGroup ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.subscribeId) +
			stringLen(this.broadcastPath) +
			stringLen(this.trackName) +
			1 + // subscriberPriority (uint8)
			1 + // subscriberOrdered (uint8)
			varintLen(this.subscriberMaxLatency) +
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

		[, err] = await writeVarint(w, this.subscribeId);
		if (err) return err;

		[, err] = await writeString(w, this.broadcastPath);
		if (err) return err;

		[, err] = await writeString(w, this.trackName);
		if (err) return err;

		[, err] = await writeUint8(w, this.subscriberPriority);
		if (err) return err;

		[, err] = await writeUint8(w, this.subscriberOrdered);
		if (err) return err;

		[, err] = await writeVarint(w, this.subscriberMaxLatency);
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

		// Read message length
		let msgLen: number;
		[msgLen, , err] = await readVarint(r);
		if (err) return err;

		// Read message body into a buffer
		const buf = new Uint8Array(msgLen);
		[, err] = await readFull(r, buf);
		if (err) return err;

		// Parse fields from the buffer
		let offset = 0;

		[this.subscribeId, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.broadcastPath, offset] = (() => {
			const [val, n] = parseString(buf, offset);
			return [val, offset + n];
		})();

		[this.trackName, offset] = (() => {
			const [val, n] = parseString(buf, offset);
			return [val, offset + n];
		})();

		[this.subscriberPriority, offset] = (() => {
			const [val, n] = parseUint8(buf, offset);
			return [val, offset + n];
		})();

		[this.subscriberOrdered, offset] = (() => {
			const [val, n] = parseUint8(buf, offset);
			return [val, offset + n];
		})();

		[this.subscriberMaxLatency, offset] = (() => {
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
