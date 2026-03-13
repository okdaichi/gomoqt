package message

import (
	"io"
)

/*
 * SUBSCRIBE_DROP Message {
 *   Subscribe ID (varint),
 *   Reason Code (varint),
 * }
 */
type SubscribeDropMessage struct {
	SubscribeID uint64
	ReasonCode  uint64
}

func (sdm SubscribeDropMessage) Len() int {
	var l int

	l += VarintLen(sdm.SubscribeID)
	l += VarintLen(sdm.ReasonCode)

	return l
}

func (sdm SubscribeDropMessage) Encode(w io.Writer) error {
	msgLen := sdm.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, sdm.SubscribeID)
	b, _ = WriteVarint(b, sdm.ReasonCode)

	_, err := w.Write(b)
	return err
}

func (sdm *SubscribeDropMessage) Decode(src io.Reader) error {
	size, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, size)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	sdm.SubscribeID = num
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	sdm.ReasonCode = num
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
