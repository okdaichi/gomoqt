/**
 * Benchmarks for GroupWriter and GroupReader frame operations.
 * 
 * Measures:
 * 1. Frame writing patterns (separate vs combined writes)
 * 2. Frame reading and buffer allocation strategies
 * 3. Memory efficiency of frame.data management
 */

import { GroupWriter, GroupReader } from "./group_stream.ts";
import { Frame } from "./frame.ts";
import { GroupMessage } from "./internal/message/mod.ts";
import { MockSendStream, MockReceiveStream } from "./mock_stream_test.ts";
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
		const frame = new Frame(new Uint8Array(1024));
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
		const frame = new Frame(new Uint8Array(4096));
		await reader.readFrame(frame);
		
		// Current: frame.data is subarray(0, 256) - holds reference to 4KB
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
		const frame = new Frame(new Uint8Array(4096));
		await reader.readFrame(frame);
		
		// Optimized: frame.data is reallocated to 256 bytes, no 4KB reference
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
		const frame = new Frame(new Uint8Array(256));
		await reader.readFrame(frame);
	},
});

Deno.bench({
	name: "Frame allocation: small (256 bytes)",
	group: "frame-allocation",
	baseline: true,
	fn: () => {
		const frame = new Frame(new Uint8Array(256));
		frame.data[0] = 1;
	},
});

Deno.bench({
	name: "Frame allocation: medium (1KB)",
	group: "frame-allocation",
	fn: () => {
		const frame = new Frame(new Uint8Array(1024));
		frame.data[0] = 1;
	},
});

Deno.bench({
	name: "Frame allocation: large (4KB)",
	group: "frame-allocation",
	fn: () => {
		const frame = new Frame(new Uint8Array(4096));
		frame.data[0] = 1;
	},
});

Deno.bench({
	name: "Frame buffer trimming: subarray (current)",
	group: "frame-trimming",
	baseline: true,
	fn: () => {
		const frame = new Frame(new Uint8Array(4096));
		// Simulate: read 256 bytes, trim buffer
		frame.data = frame.data.subarray(0, 256);
		frame.data[0] = 1;
	},
});

Deno.bench({
	name: "Frame buffer trimming: reallocate (optimized)",
	group: "frame-trimming",
	fn: () => {
		const frame = new Frame(new Uint8Array(4096));
		// Simulate: read 256 bytes, reallocate if waste >50%
		if (256 < 512 && 4096 > 256 * 2) {
			const trimmed = new Uint8Array(256);
			trimmed.set(frame.data.subarray(0, 256));
			frame.data = trimmed;
		} else {
			frame.data = frame.data.subarray(0, 256);
		}
		frame.data[0] = 1;
	},
});
