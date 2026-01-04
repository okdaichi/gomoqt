/**
 * Benchmarks for GroupWriter and GroupReader frame operations.
 *
 * Measures:
 * 1. Frame writing patterns (separate vs combined writes)
 * 2. Frame reading and buffer allocation strategies
 * 3. Memory efficiency of frame.data management
 */

import { GroupReader, GroupWriter } from "./group_stream.ts";
import { Frame } from "./frame.ts";
import { GroupMessage } from "./internal/message/mod.ts";
import { MockReceiveStream, MockSendStream } from "./mock_stream_test.ts";
import { background } from "@okdaichi/golikejs/context";

Deno.bench({
	name: "GroupWriter: writeFrame with current implementation",
	group: "groupwriter",
	baseline: true,
	fn: async () => {
		const written: Uint8Array[] = [];
		const mockStream = new MockSendStream({
			write: async (p: Uint8Array) => {
				written.push(new Uint8Array(p)); // Copy to prevent reference issues
				return [p.length, undefined];
			},
		});

		const ctx = background();
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const writer = new GroupWriter(ctx, mockStream, msg);

		// Write a 1KB frame
		const data = new Uint8Array(1024);
		const frame = new Frame(data.buffer);
		frame.write(data);
		await writer.writeFrame(frame);
	},
});

Deno.bench({
	name: "GroupReader: readFrame into pre-allocated buffer (small)",
	group: "groupreader",
	baseline: true,
	fn: async () => {
		let readCount = 0;
		const mockStream = new MockReceiveStream({
			read: async (p: Uint8Array) => {
				if (readCount === 0) {
					// First read: varint length (256 bytes)
					p[0] = 0x41; // varint: 256
					p[1] = 0x00;
					readCount++;
					return [2, undefined];
				} else {
					// Second read: actual data (256 bytes)
					const len = Math.min(p.length, 256);
					readCount++;
					return [len, undefined];
				}
			},
		});

		const ctx = background();
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const reader = new GroupReader(ctx, mockStream, msg);

		// Pre-allocate 4KB buffer, but frame is only 256 bytes
		const buffer = new ArrayBuffer(4096);
		const frame = new Frame(buffer);
		await reader.readFrame(frame);

		// Current: frame internal buffer resized to 256 bytes
	},
});

Deno.bench({
	name: "GroupReader: readFrame into pre-allocated buffer (optimized)",
	group: "groupreader-optimized",
	fn: async () => {
		let readCount = 0;
		const mockStream = new MockReceiveStream({
			read: async (p: Uint8Array) => {
				if (readCount === 0) {
					// First read: varint length (256 bytes)
					p[0] = 0x41; // varint: 256
					p[1] = 0x00;
					readCount++;
					return [2, undefined];
				} else {
					// Second read: actual data (256 bytes)
					const len = Math.min(p.length, 256);
					readCount++;
					return [len, undefined];
				}
			},
		});

		const ctx = background();
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const reader = new GroupReader(ctx, mockStream, msg);

		// Pre-allocate 4KB buffer, but frame is only 256 bytes
		const buffer = new ArrayBuffer(4096);
		const frame = new Frame(buffer);
		await reader.readFrame(frame);

		// Optimized: frame internal buffer reallocated to 256 bytes
	},
});

Deno.bench({
	name: "GroupReader: readFrame into exact-sized buffer",
	group: "groupreader",
	fn: async () => {
		let readCount = 0;
		const mockStream = new MockReceiveStream({
			read: async (p: Uint8Array) => {
				if (readCount === 0) {
					p[0] = 0x41; // 256 bytes
					p[1] = 0x00;
					readCount++;
					return [2, undefined];
				} else {
					const len = Math.min(p.length, 256);
					readCount++;
					return [len, undefined];
				}
			},
		});

		const ctx = background();
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const reader = new GroupReader(ctx, mockStream, msg);

		// Exact size buffer
		const buffer = new ArrayBuffer(256);
		const frame = new Frame(buffer);
		await reader.readFrame(frame);
	},
});

Deno.bench({
	name: "Frame allocation: small (256 bytes)",
	group: "frame-allocation",
	baseline: true,
	fn: () => {
		const data = new Uint8Array(256);
		const frame = new Frame(data.buffer);
		data[0] = 1;
		frame.write(data);
	},
});

Deno.bench({
	name: "Frame allocation: medium (1KB)",
	group: "frame-allocation",
	fn: () => {
		const data = new Uint8Array(1024);
		const frame = new Frame(data.buffer);
		data[0] = 1;
		frame.write(data);
	},
});

Deno.bench({
	name: "Frame allocation: large (4KB)",
	group: "frame-allocation",
	fn: () => {
		const data = new Uint8Array(4096);
		const frame = new Frame(data.buffer);
		data[0] = 1;
		frame.write(data);
	},
});

Deno.bench({
	name: "Frame buffer management: write small data to large buffer",
	group: "frame-trimming",
	baseline: true,
	fn: () => {
		const buffer = new ArrayBuffer(4096);
		const frame = new Frame(buffer);
		// Simulate: write 256 bytes to 4KB buffer
		const data = new Uint8Array(256);
		data[0] = 1;
		frame.write(data);
	},
});

Deno.bench({
	name: "Frame buffer management: write to exact-sized buffer",
	group: "frame-trimming",
	fn: () => {
		// Simulate: write 256 bytes to exactly-sized buffer
		const data = new Uint8Array(256);
		data[0] = 1;
		const buffer = new ArrayBuffer(256);
		const frame = new Frame(buffer);
		frame.write(data);
	},
});
