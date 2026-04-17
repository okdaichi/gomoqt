import { assertEquals, assertRejects, assertThrows } from "@std/assert";
import { Broadcast, DefaultCatalogTrackName } from "./mod.ts";
import type { TrackHandler } from "../track_mux.ts";
import type { TrackWriter } from "../track_writer.ts";

class MockTrackHandler implements TrackHandler {
	calls: TrackWriter[] = [];

	async serveTrack(trackWriter: TrackWriter): Promise<void> {
		this.calls.push(trackWriter);
	}
}

function newCatalogTrackWriter(trackName: string) {
	const frames: Uint8Array[] = [];
	const closeWithErrorCalls: number[] = [];
	let closeCalls = 0;
	let groupCloseCalls = 0;
	const group = {
		async writeFrame(payload: Uint8Array): Promise<Error | undefined> {
			frames.push(payload);
			return undefined;
		},
		async close(): Promise<void> {
			groupCloseCalls++;
		},
		async cancel(): Promise<void> {
			groupCloseCalls++;
		},
	};
	const trackWriter = {
		trackName,
		async openGroup() {
			return [group, undefined] as const;
		},
		async closeWithError(code: number): Promise<void> {
			closeWithErrorCalls.push(code);
		},
		async close(): Promise<void> {
			closeCalls++;
		},
	} as unknown as TrackWriter & {
		frames: Uint8Array[];
		closeWithErrorCalls: number[];
		getCloseCalls: () => number;
		getGroupCloseCalls: () => number;
	};
	trackWriter.frames = frames;
	trackWriter.closeWithErrorCalls = closeWithErrorCalls;
	trackWriter.getCloseCalls = () => closeCalls;
	trackWriter.getGroupCloseCalls = () => groupCloseCalls;
	return trackWriter;
}

Deno.test("msf Broadcast registerTrack updates catalog and handler", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const handler = new MockTrackHandler();

	await broadcast.registerTrack(
		{ name: "video", packaging: "loc", isLive: false },
		handler,
	);

	assertEquals(broadcast.catalog().tracks.length, 1);
	assertEquals(broadcast.catalog().tracks[0]?.name, "video");

	const trackWriter = { trackName: "video" } as TrackWriter;
	await broadcast.handler("video").serveTrack(trackWriter);
	assertEquals(handler.calls.length, 1);
	assertEquals(handler.calls[0], trackWriter);
});

Deno.test("msf Broadcast serves catalog on reserved track", async () => {
	const broadcast = new Broadcast({
		version: 1,
		tracks: [{ name: "video", packaging: "loc", isLive: false }],
	});
	const trackWriter = newCatalogTrackWriter(DefaultCatalogTrackName);

	await broadcast.serveTrack(trackWriter);

	assertEquals(trackWriter.frames.length, 1);
	const text = new TextDecoder().decode(trackWriter.frames[0]);
	assertEquals(text.includes('"video"'), true);
	assertEquals(trackWriter.getGroupCloseCalls(), 1);
	assertEquals(trackWriter.getCloseCalls(), 1);
	assertEquals(trackWriter.closeWithErrorCalls.length, 0);
});

Deno.test("msf Broadcast rejects duplicate track names across namespaces", () => {
	assertThrows(
		() =>
			new Broadcast({
				version: 1,
				defaultNamespace: "live/main",
				tracks: [
					{ name: "video", packaging: "loc", isLive: false },
					{
						namespace: "live/backup",
						name: "video",
						packaging: "loc",
						isLive: false,
					},
				],
			}),
		Error,
		"unique track names across namespaces",
	);
});

Deno.test("msf Broadcast catalogTrackName returns custom name", () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] }, "meta");
	assertEquals(broadcast.catalogTrackName, "meta");
});

Deno.test("msf Broadcast registerTrack rejects nil handler", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	await assertRejects(
		() =>
			broadcast.registerTrack(
				{ name: "video", packaging: "cmaf" },
				null as unknown as TrackHandler,
			),
		Error,
		"cannot be nil",
	);
});

Deno.test("msf Broadcast registerTrack rejects missing track name", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const handler = new MockTrackHandler();
	await assertRejects(
		() => broadcast.registerTrack({ name: "", packaging: "cmaf" }, handler),
		Error,
		"track name is required",
	);
});

Deno.test("msf Broadcast registerTrack rejects catalog track name", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const handler = new MockTrackHandler();
	await assertRejects(
		() =>
			broadcast.registerTrack({ name: DefaultCatalogTrackName, packaging: "cmaf" }, handler),
		Error,
		"reserved for the catalog track",
	);
});

Deno.test("msf Broadcast registerTrack replaces existing track", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [{ name: "audio", packaging: "cmaf" }] });
	const handler1 = new MockTrackHandler();
	await broadcast.registerTrack({ name: "audio", packaging: "cmaf", bitrate: 128 }, handler1);
	assertEquals(broadcast.catalog().tracks.length, 1);
	assertEquals(broadcast.catalog().tracks[0]?.bitrate, 128);
});

Deno.test("msf Broadcast setCatalog removes stale tracks", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [{ name: "video", packaging: "cmaf" }] });
	const handler = new MockTrackHandler();
	await broadcast.registerTrack({ name: "video", packaging: "cmaf" }, handler);

	await broadcast.setCatalog({ version: 1, tracks: [{ name: "audio", packaging: "cmaf" }] });
	assertEquals(broadcast.catalog().tracks.length, 1);
	assertEquals(broadcast.catalog().tracks[0]?.name, "audio");
});

Deno.test("msf Broadcast setCatalog with empty previous tracks", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	await broadcast.setCatalog({ version: 1, tracks: [{ name: "video", packaging: "cmaf" }] });
	assertEquals(broadcast.catalog().tracks.length, 1);
});

Deno.test("msf Broadcast removeTrack removes registered track", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [{ name: "video", packaging: "cmaf" }] });
	const handler = new MockTrackHandler();
	await broadcast.registerTrack({ name: "video", packaging: "cmaf" }, handler);

	const removed = await broadcast.removeTrack("video");
	assertEquals(removed, true);
	assertEquals(broadcast.catalog().tracks.length, 0);
});

Deno.test("msf Broadcast removeTrack returns false for empty name", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const removed = await broadcast.removeTrack("");
	assertEquals(removed, false);
});

Deno.test("msf Broadcast removeTrack returns false for catalog track name", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const removed = await broadcast.removeTrack(DefaultCatalogTrackName);
	assertEquals(removed, false);
});

Deno.test("msf Broadcast removeTrack returns true when track not in catalog but was registered", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	const handler = new MockTrackHandler();
	await broadcast.registerTrack({ name: "audio", packaging: "cmaf" }, handler);

	// Remove from catalog first via setCatalog, then removeTrack
	const removed = await broadcast.removeTrack("nonexistent");
	assertEquals(removed, false);
});

Deno.test("msf Broadcast close does not throw", async () => {
	const broadcast = new Broadcast({ version: 1, tracks: [] });
	await broadcast.close();
});

Deno.test("msf Broadcast serveCatalogTrack handles openGroup error", async () => {
	const broadcast = new Broadcast({
		version: 1,
		tracks: [{ name: "video", packaging: "cmaf" }],
	});
	const closeWithErrorCalls: number[] = [];
	const trackWriter = {
		trackName: DefaultCatalogTrackName,
		async openGroup() {
			return [undefined, new Error("open failed")] as const;
		},
		async closeWithError(code: number): Promise<void> {
			closeWithErrorCalls.push(code);
		},
		async close(): Promise<void> {},
	} as unknown as TrackWriter;

	await broadcast.serveTrack(trackWriter);
	assertEquals(closeWithErrorCalls.length, 1);
});

Deno.test("msf Broadcast serveCatalogTrack handles writeFrame error", async () => {
	const broadcast = new Broadcast({
		version: 1,
		tracks: [{ name: "video", packaging: "cmaf" }],
	});
	const closeWithErrorCalls: number[] = [];
	const cancelCalls: number[] = [];
	const group = {
		async writeFrame(_payload: Uint8Array): Promise<Error> {
			return new Error("write failed");
		},
		async cancel(code: number): Promise<void> {
			cancelCalls.push(code);
		},
		async close(): Promise<void> {},
	};
	const trackWriter = {
		trackName: DefaultCatalogTrackName,
		async openGroup() {
			return [group, undefined] as const;
		},
		async closeWithError(code: number): Promise<void> {
			closeWithErrorCalls.push(code);
		},
		async close(): Promise<void> {},
	} as unknown as TrackWriter;

	await broadcast.serveTrack(trackWriter);
	assertEquals(cancelCalls.length, 1);
	assertEquals(closeWithErrorCalls.length, 1);
});
