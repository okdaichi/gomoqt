import { assertEquals, assertThrows } from "@std/assert";
import {
	decodeLocation,
	decodeMediaTimelineEntry,
	encodeLocation,
	encodeMediaTimelineEntry,
	parseEventTimelineRecord,
	validateEventTimelineRecord,
} from "./mod.ts";

// ─── decodeLocation ──────────────────────────────────────────────────────────

Deno.test("decodeLocation rejects non-array", () => {
	assertThrows(() => decodeLocation("not-an-array"), Error, "exactly 2 items");
});

Deno.test("decodeLocation rejects wrong length", () => {
	assertThrows(() => decodeLocation([1, 2, 3]), Error, "exactly 2 items");
});

Deno.test("decodeLocation rejects non-numeric items", () => {
	assertThrows(() => decodeLocation([1, "x"]), Error, "must be numbers");
});

Deno.test("decodeLocation returns Location for valid input", () => {
	assertEquals(decodeLocation([5, 9]), { groupId: 5, objectId: 9 });
	assertEquals(decodeLocation([0, 0]), { groupId: 0, objectId: 0 });
});

// ─── encodeLocation ──────────────────────────────────────────────────────────

Deno.test("encodeLocation returns tuple", () => {
	assertEquals(encodeLocation({ groupId: 3, objectId: 7 }), [3, 7]);
});

// ─── decodeMediaTimelineEntry ─────────────────────────────────────────────────

Deno.test("decodeMediaTimelineEntry rejects invalid shape", () => {
	assertThrows(
		() => decodeMediaTimelineEntry("not-an-array"),
		Error,
		"mediaTime",
	);
});

Deno.test("decodeMediaTimelineEntry rejects wrong element types", () => {
	assertThrows(
		() => decodeMediaTimelineEntry([1, [0, "bad"], 2]),
		Error,
		"mediaTime",
	);
});

Deno.test("decodeMediaTimelineEntry returns MediaTimelineEntry for valid input", () => {
	const entry = decodeMediaTimelineEntry([100, [5, 9], 200]);
	assertEquals(entry.mediaTime, 100);
	assertEquals(entry.location, { groupId: 5, objectId: 9 });
	assertEquals(entry.wallclock, 200);
});

// ─── encodeMediaTimelineEntry ─────────────────────────────────────────────────

Deno.test("encodeMediaTimelineEntry returns tuple", () => {
	const encoded = encodeMediaTimelineEntry({
		mediaTime: 100,
		location: { groupId: 5, objectId: 9 },
		wallclock: 200,
	});
	assertEquals(encoded, [100, [5, 9], 200]);
});

// ─── validateEventTimelineRecord ─────────────────────────────────────────────

Deno.test("validateEventTimelineRecord rejects record with no timing field", () => {
	assertThrows(
		() => validateEventTimelineRecord({ data: "payload" }),
		Error,
		"exactly one of t, l, or m",
	);
});

Deno.test("validateEventTimelineRecord rejects record with multiple timing fields", () => {
	assertThrows(
		() => validateEventTimelineRecord({ t: 1, l: { groupId: 0, objectId: 0 }, data: "x" }),
		Error,
		"exactly one of t, l, or m",
	);
});

Deno.test("validateEventTimelineRecord rejects record missing data", () => {
	assertThrows(
		() => validateEventTimelineRecord({ t: 1 }),
		Error,
		"must contain data",
	);
});

Deno.test("validateEventTimelineRecord accepts record with t and data", () => {
	validateEventTimelineRecord({ t: 1, data: "payload" });
});

Deno.test("validateEventTimelineRecord accepts record with l and data", () => {
	validateEventTimelineRecord({ l: { groupId: 0, objectId: 1 }, data: 42 });
});

Deno.test("validateEventTimelineRecord accepts record with m and data", () => {
	validateEventTimelineRecord({ m: 500, data: { key: "value" } });
});

// ─── parseEventTimelineRecord ─────────────────────────────────────────────────

Deno.test("parseEventTimelineRecord rejects non-object", () => {
	assertThrows(
		() => parseEventTimelineRecord("not-an-object"),
		Error,
		"must be a JSON object",
	);
});

Deno.test("parseEventTimelineRecord returns record with t", () => {
	const record = parseEventTimelineRecord({ t: 42, data: "payload" });
	assertEquals(record.t, 42);
	assertEquals(record.data, "payload");
	assertEquals(record.l, undefined);
	assertEquals(record.m, undefined);
});

Deno.test("parseEventTimelineRecord returns record with l", () => {
	const record = parseEventTimelineRecord({ l: [3, 7], data: null });
	assertEquals(record.l, { groupId: 3, objectId: 7 });
	assertEquals(record.data, null);
});

Deno.test("parseEventTimelineRecord returns record with m", () => {
	const record = parseEventTimelineRecord({ m: 100, data: [1, 2, 3] });
	assertEquals(record.m, 100);
});

Deno.test("parseEventTimelineRecord preserves extraFields", () => {
	const record = parseEventTimelineRecord({ t: 1, data: "x", custom: "hello", num: 99 });
	assertEquals(record.extraFields, { custom: "hello", num: 99 });
});

Deno.test("parseEventTimelineRecord returns no extraFields when none present", () => {
	const record = parseEventTimelineRecord({ t: 1, data: "x" });
	assertEquals(record.extraFields, undefined);
});
