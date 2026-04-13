import type { BroadcastPath } from "./broadcast_path.ts";
import type { GroupSequence, TrackName, TrackPriority } from "./alias.ts";
import type { GroupWriter } from "./group_stream.ts";

export interface FetchRequestInit {
	broadcastPath: BroadcastPath;
	trackName: TrackName;
	priority: TrackPriority;
	groupSequence: GroupSequence;
	done?: Promise<void>;
}

export class FetchRequest {
	readonly broadcastPath: BroadcastPath;
	readonly trackName: TrackName;
	readonly priority: TrackPriority;
	readonly groupSequence: GroupSequence;
	#done: Promise<void>;

	constructor(init: FetchRequestInit) {
		this.broadcastPath = init.broadcastPath;
		this.trackName = init.trackName;
		this.priority = init.priority;
		this.groupSequence = init.groupSequence;
		this.#done = init.done ?? new Promise(() => {});
	}

	done(): Promise<void> {
		return this.#done;
	}

	withDone(done: Promise<void>): FetchRequest {
		return new FetchRequest({
			broadcastPath: this.broadcastPath,
			trackName: this.trackName,
			priority: this.priority,
			groupSequence: this.groupSequence,
			done,
		});
	}

	clone(done: Promise<void>): FetchRequest {
		return new FetchRequest({
			broadcastPath: this.broadcastPath,
			trackName: this.trackName,
			priority: this.priority,
			groupSequence: this.groupSequence,
			done,
		});
	}
}

export interface FetchHandler {
	serveFetch(w: GroupWriter, r: FetchRequest): void | Promise<void>;
}
