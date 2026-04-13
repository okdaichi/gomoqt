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

export interface FetchMessageInit {
	broadcastPath?: string;
	trackName?: string;
	priority?: number;
	groupSequence?: number;
}

export class FetchMessage {
	broadcastPath: string;
	trackName: string;
	priority: number;
	groupSequence: number;

	constructor(init: FetchMessageInit = {}) {
		this.broadcastPath = init.broadcastPath ?? "";
		this.trackName = init.trackName ?? "";
		this.priority = init.priority ?? 0;
		this.groupSequence = init.groupSequence ?? 0;
	}

	get len(): number {
		return (
			stringLen(this.broadcastPath) +
			stringLen(this.trackName) +
			varintLen(this.priority) +
			varintLen(this.groupSequence)
		);
	}

	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeVarint(w, msgLen);
		if (err) return err;

		[, err] = await writeString(w, this.broadcastPath);
		if (err) return err;

		[, err] = await writeString(w, this.trackName);
		if (err) return err;

		[, err] = await writeVarint(w, this.priority);
		if (err) return err;

		[, err] = await writeVarint(w, this.groupSequence);
		if (err) return err;

		return undefined;
	}

	async decode(r: Reader): Promise<Error | undefined> {
		let err: Error | undefined;

		let msgLen: number;
		[msgLen, , err] = await readVarint(r);
		if (err) return err;

		const buf = new Uint8Array(msgLen);
		[, err] = await readFull(r, buf);
		if (err) return err;

		let offset = 0;

		[this.broadcastPath, offset] = (() => {
			const [str, n] = parseString(buf, offset);
			return [str, offset + n];
		})();

		[this.trackName, offset] = (() => {
			const [str, n] = parseString(buf, offset);
			return [str, offset + n];
		})();

		[this.priority, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.groupSequence, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		if (offset !== msgLen) {
			return new Error("message length mismatch");
		}

		return undefined;
	}
}
