package moqt

type FetchRequest struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName
	Priority      TrackPriority
	GroupSequence GroupSequence
}

type FetchHandler interface {
	ServeFetch(w *GroupWriter, r *FetchRequest)
}

type FetchHandlerFunc func(w *GroupWriter, r *FetchRequest)

func (f FetchHandlerFunc) ServeFetch(w *GroupWriter, r *FetchRequest) {
	f(w, r)
}
