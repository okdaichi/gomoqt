package moqt

import (
	"sync"

	"github.com/okdaichi/gomoqt/moqt/internal/message"
	"github.com/okdaichi/gomoqt/transport"
)

func newReceiveSubscribeStream(id SubscribeID, stream transport.Stream, config *SubscribeConfig) *receiveSubscribeStream {
	substr := &receiveSubscribeStream{
		subscribeID: id,
		config:      config,
		stream:      stream,
		updatedCh:   make(chan struct{}, 1),
	}

	// Listen for updates in a separate goroutine
	go func() {
		var updateMsg message.SubscribeUpdateMessage
		var err error

		for {
			err = updateMsg.Decode(substr.stream)
			if err != nil {
				break
			}

			config := &SubscribeConfig{
				Priority:   TrackPriority(updateMsg.SubscriberPriority),
				Ordered:    boolFromWireFlag(updateMsg.SubscriberOrdered),
				MaxLatency: updateMsg.SubscriberMaxLatency,
				StartGroup: groupSequenceFromWire(updateMsg.StartGroup),
				EndGroup:   groupSequenceFromWire(updateMsg.EndGroup),
			}

			substr.mu.Lock()

			substr.config = config
			select {
			case substr.updatedCh <- struct{}{}:
			default:
			}
			substr.mu.Unlock()
		}

	}()

	return substr
}

type receiveSubscribeStream struct {
	subscribeID SubscribeID

	stream transport.Stream

	acceptOnce sync.Once
	// writeInfoWG tracks active WriteInfo calls so close waits for them.
	writeInfoWG sync.WaitGroup

	mu        sync.Mutex
	config    *SubscribeConfig
	updatedCh chan struct{}
}

func (substr *receiveSubscribeStream) SubscribeID() SubscribeID {
	return substr.subscribeID
}

func (substr *receiveSubscribeStream) writeInfo(info PublishInfo) error {
	var err error
	substr.acceptOnce.Do(func() {
		substr.writeInfoWG.Add(1)
		defer substr.writeInfoWG.Done()

		substr.mu.Lock()
		defer substr.mu.Unlock()

		ordered := boolToWireFlag(info.Ordered)

		startGroup := groupSequenceToWire(info.StartGroup)

		endGroup := groupSequenceToWire(info.EndGroup)

		sum := message.SubscribeOkMessage{
			PublisherPriority:   uint8(info.Priority),
			PublisherOrdered:    ordered,
			PublisherMaxLatency: info.MaxLatency,
			StartGroup:          startGroup,
			EndGroup:            endGroup,
		}
		err = sum.Encode(substr.stream)
		if err != nil {
			_ = substr.closeWithError(SubscribeErrorCodeInternal)
			return
		}
	})

	return err
}

func (substr *receiveSubscribeStream) TrackConfig() *SubscribeConfig {
	substr.mu.Lock()
	defer substr.mu.Unlock()

	// Ensure config is never nil
	if substr.config == nil {
		substr.config = &SubscribeConfig{}
	}

	return substr.config
}

func (substr *receiveSubscribeStream) Updated() <-chan struct{} {
	return substr.updatedCh
}

func (substr *receiveSubscribeStream) close() error {
	substr.mu.Lock()
	defer substr.mu.Unlock()

	if updateCh := substr.updatedCh; updateCh != nil {
		substr.updatedCh = nil
		close(updateCh)
	}

	return substr.stream.Close()
}

func (substr *receiveSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	substr.mu.Lock()
	defer substr.mu.Unlock()

	strErrCode := transport.StreamErrorCode(code)
	cancelStreamWithError(substr.stream, strErrCode)

	if updateCh := substr.updatedCh; updateCh != nil {
		substr.updatedCh = nil
		close(updateCh)
	}

	return nil
}
