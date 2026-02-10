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

	get bytes(): Uint8Array {
		return new Uint8Array(this.#buf, 0, this.#len);
	}

	constructor(buffer?: ArrayBuffer | Uint8Array) {
		if (buffer instanceof Uint8Array) {
			const slice = buffer.buffer.slice(
				buffer.byteOffset,
				buffer.byteOffset + buffer.byteLength,
			);
			this.#buf = slice instanceof SharedArrayBuffer
				? new ArrayBuffer(buffer.byteLength)
				: slice;
			if (slice instanceof SharedArrayBuffer) {
				new Uint8Array(this.#buf).set(buffer);
			}
			this.#len = buffer.byteLength;
		} else {
			this.#buf = buffer ?? new ArrayBuffer(0);
		}
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

		if (target.byteLength < this.#len) {
			throw new Error(
				`Destination buffer too small: ${target.byteLength} < ${this.#len}`,
			);
		}

		// Ensure we don't exceed the buffer bounds
		const copyLen = Math.min(this.#len, this.#buf.byteLength);
		target.set(new Uint8Array(this.#buf, 0, copyLen));
	}
}

export interface Frame extends ByteSource, ByteSink {
	readonly bytes: Uint8Array;
}

export const Frame: {
	new (buffer: ArrayBuffer | Uint8Array): Frame;
} = BytesBuffer;
