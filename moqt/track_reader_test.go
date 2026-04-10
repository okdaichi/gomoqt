package moqt

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/okdaichi/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrackReader(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	substr := newSendSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{}, nil)
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	assert.NotNil(t, receiver, "newTrackReader should not return nil")
	require.NotNil(t, receiver.Request)
	assert.Equal(t, BroadcastPath("/test"), receiver.Request.BroadcastPath)
	assert.Equal(t, TrackName("video"), receiver.Request.TrackName)
	// Verify info propagation
	assert.Equal(t, PublishInfo{}, substr.ReadInfo(), "sendSubscribeStream should return the Info passed at construction")
	assert.NotNil(t, receiver.queueing, "queue should be initialized")
	assert.NotNil(t, receiver.queuedCh, "queuedCh should be initialized")
	assert.NotNil(t, receiver.dequeued, "dequeued should be initialized")
}

func TestTrackReader_AcceptGroup(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	// Test with a timeout to ensure we don't block forever when no groups are available
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

func TestTrackReader_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	// Cancel the context
	cancel()

	// Test that AcceptGroup returns context error when context is cancelled
	testCtx, testCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer testCancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected error when context is cancelled")
	// Should return context.Canceled or DeadlineExceeded
	assert.True(t, err == context.Canceled || err == context.DeadlineExceeded, "expected context error")
}

func TestTrackReader_Context_FollowsStreamLifecycle(t *testing.T) {
	streamCtx := context.Background()
	_, cancelSetup := context.WithCancel(context.Background())
	defer cancelSetup()

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(streamCtx).Maybe()
	mockStream.On("Close").Return(nil).Maybe()

	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	// Cancel setup context; TrackReader context should remain alive while stream is alive.
	cancelSetup()
	select {
	case <-receiver.Context().Done():
		t.Fatal("track reader context should not be canceled by request setup context")
	case <-time.After(20 * time.Millisecond):
		// expected
	}

	// Close stream; TrackReader context should be canceled.
	require.NoError(t, mockStream.Close())

	select {
	case <-receiver.Context().Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("track reader context should be canceled when stream is closed")
	}
}

func TestTrackReader_EnqueueGroup(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	// Mock receive stream
	mockReceiveStream := &MockQUICReceiveStream{}

	// Enqueue a group
	receiver.enqueueGroup(GroupSequence(1), mockReceiveStream)

	// Test that we can accept the enqueued group
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	group, err := receiver.AcceptGroup(testCtx)
	assert.NoError(t, err, "should be able to accept enqueued group")
	assert.NotNil(t, group, "accepted group should not be nil")

	mockReceiveStream.AssertExpectations(t)
}

func TestTrackReader_AcceptGroup_RealImplementation(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	// Test with a timeout to ensure we don't block forever
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

func TestTrackReader_Close(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.Anything).Return(nil)
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	err := receiver.Close()
	assert.NoError(t, err)

	// Close again should not error
	err = receiver.Close()
	assert.NoError(t, err)
}

func TestTrackReader_Update(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	newTrackConfig := SubscribeConfig{}

	_ = receiver.Update(&newTrackConfig)

	// Verify update
	assert.Equal(t, &SubscribeConfig{}, receiver.TrackConfig())
}

func TestTrackReader_HandleDrop(t *testing.T) {
	var buf bytes.Buffer
	_, _ = buf.Write([]byte{byte(message.MessageTypeSubscribeDrop)})
	require.NoError(t, (message.SubscribeDropMessage{
		StartGroup: 11,
		EndGroup:   21,
		ErrorCode:  3,
	}).Encode(&buf))

	mockStream := &MockQUICStream{
		ReadFunc: buf.Read,
	}
	mockStream.On("Context").Return(context.Background()).Maybe()

	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	done := make(chan SubscribeDrop, 1)
	receiver.HandleDrop(func(drop SubscribeDrop) {
		done <- drop
	})

	select {
	case drop := <-done:
		assert.Equal(t, SubscribeDrop{
			StartGroup: 10,
			EndGroup:   20,
			ErrorCode:  3,
		}, drop)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected drop callback to be invoked")
	}
}

func TestTrackReader_CloseWithError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.Anything).Return(nil)
	mockStream.On("CancelWrite", mock.Anything).Return(nil)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	receiver.CloseWithError(SubscribeErrorCodeInternal)
}

func TestGroupReader_CancelRead_RemovesFromManager(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	substr := newTestSendSubscribeStream(mockStream, &SubscribeConfig{})
	receiver := newTrackReader(testSubscribeRequest(t, nil), substr, func() {})

	recvStream := &MockQUICReceiveStream{}
	recvStream.On("CancelRead", mock.Anything).Return()
	group := newGroupReader(GroupSequence(1), recvStream, receiver.groupManager)

	assert.Len(t, receiver.groupManager.activeGroups, 1)
	assert.Contains(t, receiver.groupManager.activeGroups, group)

	group.CancelRead(InternalGroupErrorCode)
	assert.Len(t, receiver.groupManager.activeGroups, 0)
	assert.NotContains(t, receiver.groupManager.activeGroups, group)
}
