export interface ByteSource {
	readonly byteLength: number;
	copyTo(target: ArrayBuffer | ArrayBufferView): void;
}

export interface ByteSink {
	write(p: Uint8Array): void | Promise<void>;
}

export type ByteSinkFunc = (p: Uint8Array) => void | Promise<void>;

export class BytesBuffer implements ByteSource, ByteSink {
	#buf: ArrayBuffer; // Internal buffer (full capacity)

	#len: number = 0; // Actual data length

	get byteLength(): number {
		return this.#len;
	}

	constructor(buffer?: ArrayBuffer) {
		this.#buf = buffer ?? new ArrayBuffer(0);
	}

	write(p: Uint8Array): void {
		if (this.#buf.byteLength < p.byteLength) {
			// Resize buffer if necessary
			this.#buf = new ArrayBuffer(p.byteLength);
		}

		const target = new Uint8Array(this.#buf, 0, p.byteLength);
		target.set(p);
		this.#len = p.byteLength;
	}

	copyTo(dest: AllowSharedBufferSource): void {
		let target: Uint8Array;
		if (dest instanceof Uint8Array) {
			target = dest;
		} else if (dest instanceof ArrayBuffer || dest instanceof SharedArrayBuffer) {
			target = new Uint8Array(dest as ArrayBuffer); // Handle both ArrayBuffer and SharedArrayBuffer
		} else {
			throw new Error("Unsupported destination type");
		}

		if (target.byteLength < this.#buf.byteLength) {
			throw new Error(
				`Destination buffer too small: ${target.byteLength} < ${this.#buf.byteLength}`,
			);
		}

		target.set(new Uint8Array(this.#buf, 0, this.#buf.byteLength));
	}
}

export interface Frame extends ByteSource, ByteSink {}

export const Frame: {
	new (buffer: ArrayBuffer): Frame;
} = BytesBuffer;
