package moqt

import (
	"context"
	"fmt"
)

type FetchRequest struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName
	Priority      TrackPriority
	GroupSequence GroupSequence

	ctx context.Context
}

// Context returns the request's context. To change the context, use
// [FetchRequest.WithContext].
//
// The returned context is always non-nil; it defaults to the
// background context.
func (r *FetchRequest) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (r *FetchRequest) WithContext(ctx context.Context) *FetchRequest {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(FetchRequest)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

// Clone returns a deep copy of r with its context changed to ctx.
// The provided ctx must be non-nil.
func (r *FetchRequest) Clone(ctx context.Context) *FetchRequest {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(FetchRequest)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

type FetchHandler interface {
	ServeFetch(w *GroupWriter, r *FetchRequest)
}

type FetchHandlerFunc func(w *GroupWriter, r *FetchRequest)

func (f FetchHandlerFunc) ServeFetch(w *GroupWriter, r *FetchRequest) {
	f(w, r)
}

func validateFetchHandler(handler FetchHandler) error {
	if handler == nil {
		return fmt.Errorf("moqt: fetch handler cannot be nil")
	}
	if f, ok := handler.(FetchHandlerFunc); ok && f == nil {
		return fmt.Errorf("moqt: fetch handler function cannot be nil")
	}
	return nil
}

func safeServeFetch(handler FetchHandler, w *GroupWriter, r *FetchRequest) (err error) {
	if e := validateFetchHandler(handler); e != nil {
		return e
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic during fetch handling: %v", p)
		}
	}()
	handler.ServeFetch(w, r)
	return nil
}
