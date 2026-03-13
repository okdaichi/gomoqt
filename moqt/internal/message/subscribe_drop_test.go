package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeDropMessage_EncodeDecode(t *testing.T) {
	t.Run("valid_message", func(t *testing.T) {
		original := SubscribeDropMessage{
			SubscribeID: 42,
			ReasonCode: 3,
		}

		var buf bytes.Buffer
		err := original.Encode(&buf)
		require.NoError(t, err)

		var decoded SubscribeDropMessage
		err = decoded.Decode(&buf)
		require.NoError(t, err)

		assert.Equal(t, original.SubscribeID, decoded.SubscribeID)
		assert.Equal(t, original.ReasonCode, decoded.ReasonCode)
	})

	t.Run("zero_subscribe_id", func(t *testing.T) {
		original := SubscribeDropMessage{
			SubscribeID: 0,
			ReasonCode: 1,
		}

		var buf bytes.Buffer
		err := original.Encode(&buf)
		require.NoError(t, err)

		var decoded SubscribeDropMessage
		err = decoded.Decode(&buf)
		require.NoError(t, err)

		assert.Equal(t, uint64(0), decoded.SubscribeID)
		assert.Equal(t, uint64(1), decoded.ReasonCode)
	})

	t.Run("max_values", func(t *testing.T) {
		// Max varint value in MOQ (62-bit limit)
		maxVal := uint64(1<<62) - 1
		original := SubscribeDropMessage{
			SubscribeID: maxVal,
			ReasonCode: maxVal,
		}

		var buf bytes.Buffer
		err := original.Encode(&buf)
		require.NoError(t, err)

		var decoded SubscribeDropMessage
		err = decoded.Decode(&buf)
		require.NoError(t, err)

		assert.Equal(t, maxVal, decoded.SubscribeID)
		assert.Equal(t, maxVal, decoded.ReasonCode)
	})
}

func TestSubscribeDropMessage_DecodeErrors(t *testing.T) {
	t.Run("read_message_length_error", func(t *testing.T) {
		var sdm SubscribeDropMessage
		var buf bytes.Buffer
		src := bytes.NewReader(buf.Bytes())
		err := sdm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read_full_error", func(t *testing.T) {
		var sdm SubscribeDropMessage
		var buf bytes.Buffer
		buf.WriteByte(0x10) // length varint = 16
		src := bytes.NewReader(buf.Bytes())
		err := sdm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read_varint_error_for_subscribe_id", func(t *testing.T) {
		var sdm SubscribeDropMessage
		var buf bytes.Buffer
		buf.WriteByte(0x01) // length varint = 1
		buf.WriteByte(0x80) // invalid varint (incomplete)
		src := bytes.NewReader(buf.Bytes())
		err := sdm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read_varint_error_for_reason_code", func(t *testing.T) {
		var sdm SubscribeDropMessage
		var buf bytes.Buffer
		buf.WriteByte(0x02) // length varint = 2
		buf.WriteByte(0x00) // SubscribeID = 0
		buf.WriteByte(0x80) // invalid varint (incomplete)
		src := bytes.NewReader(buf.Bytes())
		err := sdm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra_data", func(t *testing.T) {
		var sdm SubscribeDropMessage
		var buf bytes.Buffer
		buf.WriteByte(0x03) // length varint = 3
		buf.WriteByte(0x01) // SubscribeID = 1
		buf.WriteByte(0x02) // ReasonCode = 2
		buf.WriteByte(0xFF) // extra trailing byte
		src := bytes.NewReader(buf.Bytes())
		err := sdm.Decode(src)
		assert.Equal(t, ErrMessageTooShort, err)
	})
}
