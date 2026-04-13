import { assertEquals } from "@std/assert";
import { SubscribeOkMessage } from "./subscribe_ok.ts";
import { Buffer } from "@okdaichi/golikejs/bytes";

Deno.test("SubscribeOkMessage - encode/decode roundtrip", async (t) => {
	const testCases = {
		"zero values": {
			publisherPriority: 0,
			publisherOrdered: 0,
			publisherMaxLatency: 0,
			startGroup: 0,
			endGroup: 0,
		},
		"normal case": {
			publisherPriority: 1,
			publisherOrdered: 1,
			publisherMaxLatency: 100,
			startGroup: 5,
			endGroup: 10,
		},
		"max priority": {
			publisherPriority: 255,
			publisherOrdered: 0,
			publisherMaxLatency: 0,
			startGroup: 0,
			endGroup: 0,
		},
		"mid priority with latency": {
			publisherPriority: 10,
			publisherOrdered: 1,
			publisherMaxLatency: 500,
			startGroup: 0,
			endGroup: 20,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const buffer = Buffer.make(100);
			const message = new SubscribeOkMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SubscribeOkMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(
				decodedMessage.publisherPriority,
				input.publisherPriority,
				`publisherPriority mismatch for ${caseName}`,
			);
			assertEquals(
				decodedMessage.publisherOrdered,
				input.publisherOrdered,
				`publisherOrdered mismatch for ${caseName}`,
			);
			assertEquals(
				decodedMessage.publisherMaxLatency,
				input.publisherMaxLatency,
				`publisherMaxLatency mismatch for ${caseName}`,
			);
			assertEquals(
				decodedMessage.startGroup,
				input.startGroup,
				`startGroup mismatch for ${caseName}`,
			);
			assertEquals(
				decodedMessage.endGroup,
				input.endGroup,
				`endGroup mismatch for ${caseName}`,
			);
		});
	}

	await t.step("decode should return error when readVarint fails", async () => {
		const buffer = Buffer.make(0); // Empty buffer causes read error
		const message = new SubscribeOkMessage({});
		const err = await message.decode(buffer);
		assertEquals(err !== undefined, true);
	});
});
