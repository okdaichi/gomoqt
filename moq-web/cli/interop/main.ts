import { parseArgs } from "@std/cli/parse-args";
import { Client } from "../../src/client.ts";
import { TrackMux } from "../../src/track_mux.ts";
import { Frame } from "../../src/frame.ts";

async function main() {
	const args = parseArgs(Deno.args, {
		string: ["addr", "cert-hash"],
		boolean: ["insecure", "debug"],
		default: { addr: "https://localhost:9000", insecure: false, debug: false },
	});

	// Suppress debug logs unless --debug flag is provided
	if (!args.debug) {
		console.debug = () => {};
	}

	const addr = args.addr;
	// helper prints a step and appends ok/failed (streamAndLog prefixes [Client])
	async function step<T>(msg: string, fn: () => Promise<T>): Promise<T> {
		process.stdout.write(`${msg}...`);
		try {
			const res = await fn();
			console.log(" ok");
			return res;
		} catch (err) {
			console.log(" failed");
			throw err;
		}
	}

	console.log(`Connecting to server at ${addr}`);

	// For local development with self-signed certificates
	let transportOptions: WebTransportOptions = {};

	if (args.insecure && args["cert-hash"]) {
		// Use provided certificate hash for localhost development
		const hashBase64 = args["cert-hash"];
		const hashBytes = Uint8Array.from(atob(hashBase64), (c) => c.charCodeAt(0));

		transportOptions = {
			serverCertificateHashes: [{
				algorithm: "sha-256",
				value: hashBytes,
			}],
		};
		console.log("[DEV] Using provided certificate hash for self-signed cert");
	}

	const client = new Client({ transportOptions });
	const mux = new TrackMux();

	// Register publish handler before dialing
	const donePromise = new Promise<void>((resolve) => {
		// Create a context promise that never resolves (unless we want to stop publishing)
		const ctx = new Promise<void>(() => {});

		// helper that prints step and status
		async function step<T>(msg: string, fn: () => Promise<T>): Promise<T> {
			process.stdout.write(`${msg}...`);
			try {
				const res = await fn();
				console.log(" ok");
				return res;
			} catch (e) {
				console.log(" failed");
				throw e;
			}
		}

		mux.publishFunc(ctx, "/interop/client", async (tw) => {
			try {
				const [group] = await step("Opening group", () => tw.openGroup());

				const data = new Uint8Array([72, 69, 76, 76, 79]); // "HELLO"
				const frame = new Frame(data.buffer);
				frame.write(data);
				await step("Writing frame to server", () => group.writeFrame(frame));

				await group.close();
				resolve();
			} catch (err) {
				console.error("Error in publish handler:", err);
				resolve(); // Resolve anyway to avoid hanging
			}
		});
	});

	let sess;
	try {
		sess = await step("Connecting to server", () => client.dial(addr, mux));
	} catch (err) {
		console.error("Failed to connect:", err);
		return;
	}

	try {
		// Step 1: Accept announcements from server
		const [anns, acceptErr] = await step("Accepting server announcements", () => sess.acceptAnnounce("/"));
		if (acceptErr) throw acceptErr;

		const [ann, annErr] = await step("Receiving announcement", () => anns.receive(new Promise(() => {})));
		if (annErr) throw annErr;
		if (!ann) {
			throw new Error("Announcement stream closed");
		}
		console.log(`Discovered broadcast: ${ann.broadcastPath}`);

		// Close announcement stream after receiving what we need
		anns.close();

		// Step 2: Subscribe to the server's broadcast and receive data
		const [track, subErr] = await step("Subscribing to server broadcast", () => sess.subscribe(ann.broadcastPath, ""));
		if (subErr) throw subErr;

		const [group, groupErr] = await step("Accepting group from server", () => track.acceptGroup(new Promise(() => {})));
		if (groupErr) throw groupErr;
		if (!group) {
			throw new Error("Track closed before group received");
		}

		const frame = new Frame(new Uint8Array(1024));
		const readErr = await step("Reading the first frame from server", () => group.readFrame(frame));
		if (readErr) throw readErr;

		// Note: frame.data might contain trailing zeros if payload is smaller than 1024
		const payload = new TextDecoder().decode(frame.bytes).replace(/\0/g, "");
		console.log(`Payload: ${payload}`);

		// Wait for the publish handler to complete
		await donePromise;

		// Wait a bit for server to process everything before closing
		await new Promise((r) => setTimeout(r, 1000));
	} catch (err) {
		// Ignore connection reset errors during shutdown
		if (err && typeof err === "object" && "message" in err) {
			const msg = String(err.message);
			if (msg.includes("ConnectionReset") || msg.includes("stream reset")) {
				// Expected during session closure
			} else {
				console.error("Error during interop:", err);
			}
		} else {
			console.error("Error during interop:", err);
		}
	} finally {
		// close using the same step helper so the status appears on one line
		if (sess) {
			await step("Closing session", () => sess.closeWithError(0, "no error"));
		}
	}
}

if (import.meta.main) {
	main().catch((err) => {
		console.error("Fatal error:", err);
		Deno.exit(1);
	});
}
