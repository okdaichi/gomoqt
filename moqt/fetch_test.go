package moqt

import (
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
		failed := safeServeFetch(nil, nil, nil)
		assert.True(t, failed)
	})

	t.Run("fails on typed nil function handler", func(t *testing.T) {
		var f FetchHandlerFunc
		var h FetchHandler = f

		failed := safeServeFetch(h, nil, nil)
		assert.True(t, failed)
	})

	t.Run("fails when handler panics", func(t *testing.T) {
		called := false
		h := FetchHandlerFunc(func(w *GroupWriter, r *FetchRequest) {
			called = true
			panic("boom")
		})

		failed := safeServeFetch(h, &GroupWriter{}, &FetchRequest{})
		assert.True(t, failed)
		assert.True(t, called)
	})

	t.Run("succeeds on normal handler", func(t *testing.T) {
		h := &testFetchHandler{}
		w := &GroupWriter{}
		r := &FetchRequest{BroadcastPath: "/ok", TrackName: "audio"}

		failed := safeServeFetch(h, w, r)
		assert.False(t, failed)
		assert.True(t, h.called)
		assert.Equal(t, w, h.gotW)
		assert.Equal(t, r, h.gotR)
	})
}
