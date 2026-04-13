import type { TrackPriority } from "./alias.ts";

export interface Info {
	priority: TrackPriority;
	ordered: boolean;
	maxLatency: number;
	startGroup: number;
	endGroup: number;
}
