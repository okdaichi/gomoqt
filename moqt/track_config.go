package moqt

import (
	"fmt"
)

// TrackConfig holds subscription parameters for a track. It is used to
// specify the delivery priority for the track.
type TrackConfig struct {
	SubscriberPriority TrackPriority
	Ordered            bool
	MaxLatency         uint64
	StartGroup         GroupSequence
	EndGroup           GroupSequence
}

func (sc TrackConfig) String() string {
	return fmt.Sprintf("{ subscriber_priority: %d, ordered: %t, max_latency_ms: %d, start_group: %d, end_group: %d }", sc.SubscriberPriority, sc.Ordered, sc.MaxLatency, sc.StartGroup, sc.EndGroup)
}
