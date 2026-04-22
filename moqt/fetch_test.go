package moqt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testFetchHandler struct {
	called bool
	gotW   *GroupWriter
	gotR   *FetchRequest
}

func (h *testFetchHandler) ServeFetch(w *GroupWriter, r *FetchRequest) {
	h.called = true
	h.gotW = w
	h.gotR = r
}

func TestFetchHandlerFunc_ServeFetch(t *testing.T) {
	called := false
	var gotW *GroupWriter
	var gotR *FetchRequest

	h := FetchHandlerFunc(func(w *GroupWriter, r *FetchRequest) {
		called = true
		gotW = w
		gotR = r
	})

	w := &GroupWriter{}
	r := &FetchRequest{BroadcastPath: "/test/path", TrackName: "video"}
	h.ServeFetch(w, r)

	assert.True(t, called)
	assert.Equal(t, w, gotW)
	assert.Equal(t, r, gotR)
}

func TestValidateFetchHandler(t *testing.T) {
	t.Run("nil handler", func(t *testing.T) {
		err := validateFetchHandler(nil)
		assert.Error(t, err)
	})

	t.Run("typed nil function handler", func(t *testing.T) {
		var f FetchHandlerFunc
		var h FetchHandler = f

		err := validateFetchHandler(h)
		assert.Error(t, err)
	})

	t.Run("valid function handler", func(t *testing.T) {
		h := FetchHandlerFunc(func(w *GroupWriter, r *FetchRequest) {})
		err := validateFetchHandler(h)
		assert.NoError(t, err)
	})

	t.Run("valid concrete handler", func(t *testing.T) {
		h := &testFetchHandler{}
		err := validateFetchHandler(h)
		assert.NoError(t, err)
	})
}

func TestSafeServeFetch(t *testing.T) {
	t.Run("fails on nil handler", func(t *testing.T) {
		err := safeServeFetch(nil, nil, nil)
		assert.NotNil(t, err)
	})

	t.Run("fails on typed nil function handler", func(t *testing.T) {
		var f FetchHandlerFunc
		var h FetchHandler = f

		err := safeServeFetch(h, nil, nil)
		assert.NotNil(t, err)
	})

	t.Run("fails when handler panics", func(t *testing.T) {
		called := false
		h := FetchHandlerFunc(func(w *GroupWriter, r *FetchRequest) {
			called = true
			panic("boom")
		})

		err := safeServeFetch(h, &GroupWriter{}, &FetchRequest{})
		assert.NotNil(t, err)
		assert.True(t, called)
	})

	t.Run("succeeds on normal handler", func(t *testing.T) {
		h := &testFetchHandler{}
		w := &GroupWriter{}
		r := &FetchRequest{BroadcastPath: "/ok", TrackName: "audio"}

		err := safeServeFetch(h, w, r)
		assert.Nil(t, err)
		assert.True(t, h.called)
		assert.Equal(t, w, h.gotW)
		assert.Equal(t, r, h.gotR)
	})
}

func TestFetchRequest_Context(t *testing.T) {
	t.Run("returns background context by default", func(t *testing.T) {
		r := &FetchRequest{}
		assert.Equal(t, context.Background(), r.Context())
	})

	t.Run("returns set context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		r := &FetchRequest{ctx: ctx}
		assert.Equal(t, ctx, r.Context())
	})
}

func TestFetchRequest_WithContext(t *testing.T) {
	t.Run("returns shallow copy with new context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		r := &FetchRequest{BroadcastPath: "/test", TrackName: "video"}
		r2 := r.WithContext(ctx)

		assert.Equal(t, ctx, r2.Context())
		assert.Equal(t, r.BroadcastPath, r2.BroadcastPath)
		assert.Equal(t, r.TrackName, r2.TrackName)
		// Original unchanged
		assert.Equal(t, context.Background(), r.Context())
	})

	t.Run("panics on nil context", func(t *testing.T) {
		r := &FetchRequest{}
		var nilCtx context.Context
		assert.Panics(t, func() { r.WithContext(nilCtx) })
	})
}

func TestFetchRequest_Clone(t *testing.T) {
	t.Run("returns deep copy with new context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		r := &FetchRequest{
			BroadcastPath: "/test",
			TrackName:     "video",
			Priority:      3,
			GroupSequence: 42,
		}
		r2 := r.Clone(ctx)

		assert.Equal(t, ctx, r2.Context())
		assert.Equal(t, r.BroadcastPath, r2.BroadcastPath)
		assert.Equal(t, r.TrackName, r2.TrackName)
		assert.Equal(t, r.Priority, r2.Priority)
		assert.Equal(t, r.GroupSequence, r2.GroupSequence)
	})

	t.Run("panics on nil context", func(t *testing.T) {
		r := &FetchRequest{}
		var nilCtx context.Context
		assert.Panics(t, func() { r.Clone(nilCtx) })
	})
}
