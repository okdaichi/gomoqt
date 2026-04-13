import type { Reader, Writer } from "@okdaichi/golikejs/io";
import {
	parseString,
	parseVarint,
	readFull,
	readVarint,
	stringLen,
	varintLen,
	writeString,
	writeVarint,
} from "./message.ts";

export interface AnnounceMessageInit {
	suffix?: string;
	active?: boolean;
	hops?: number;
}

export class AnnounceMessage {
	suffix: string;
	active: boolean;
	hops: number;

	constructor(init: AnnounceMessageInit = {}) {
		this.suffix = init.suffix ?? "";
		this.active = init.active ?? false;
		this.hops = init.hops ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.active ? 1 : 0) +
			stringLen(this.suffix) +
			varintLen(this.hops)
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

		// Write AnnounceStatus as varint: 0x0 (ENDED) or 0x1 (ACTIVE)
		[, err] = await writeVarint(w, this.active ? 1 : 0);
		if (err) return err;

		[, err] = await writeString(w, this.suffix);
		if (err) return err;

		[, err] = await writeVarint(w, this.hops);
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

		// Read AnnounceStatus as varint
		const [status, n1] = parseVarint(buf, offset);
		this.active = status === 1;
		offset += n1;

		[this.suffix, offset] = (() => {
			const [val, n] = parseString(buf, offset);
			return [val, offset + n];
		})();

		[this.hops] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		return undefined;
	}
}
