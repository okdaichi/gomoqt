import { assertEquals } from "@std/assert";
import type { ConnectInit, MOQOptions } from "./options.ts";

// Test configuration to ignore resource leaks from background operations
const testOptions = {
	sanitizeResources: false,
	sanitizeOps: false,
};

Deno.test("MOQOptions", testOptions, async (t) => {
	await t.step("should allow empty options", () => {
		// All options should be optional
		const emptyOptions: MOQOptions = {};
		assertEquals(emptyOptions.transportOptions, undefined);
	});

	await t.step("should support transportOptions", () => {
		const transportOptions: WebTransportOptions = {
			allowPooling: true,
			congestionControl: "throughput",
		};

		const options: MOQOptions = {
			transportOptions: transportOptions,
		};

		assertEquals(options.transportOptions, transportOptions);
		assertEquals(options.transportOptions?.allowPooling, true);
		assertEquals(options.transportOptions?.congestionControl, "throughput");
	});
});

Deno.test("ConnectInit", testOptions, async (t) => {
	await t.step("should allow empty init", () => {
		const emptyInit: ConnectInit = {};
		assertEquals(emptyInit.transportOptions, undefined);
		assertEquals(emptyInit.mux, undefined);
		assertEquals(emptyInit.onGoaway, undefined);
		assertEquals(emptyInit.transportFactory, undefined);
	});

	await t.step("should support transportOptions", () => {
		const transportOptions: WebTransportOptions = {
			allowPooling: true,
			congestionControl: "throughput",
		};

		const init: ConnectInit = {
			transportOptions: transportOptions,
		};

		assertEquals(init.transportOptions, transportOptions);
		assertEquals(init.transportOptions?.allowPooling, true);
		assertEquals(init.transportOptions?.congestionControl, "throughput");
	});
});
