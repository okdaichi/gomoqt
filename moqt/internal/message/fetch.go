package message

import "io"

type FetchMessage struct {
	BroadcastPath string
	TrackName     string
	Priority      uint8
	GroupSequence uint64
}

func (f FetchMessage) Len() int {
	var l int
	l += StringLen(f.BroadcastPath)
	l += StringLen(f.TrackName)
	l += VarintLen(uint64(f.Priority))
	l += VarintLen(f.GroupSequence)
	return l
}

func (f FetchMessage) Encode(w io.Writer) error {
	msgLen := f.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))
	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(len(f.BroadcastPath)))
	b = append(b, f.BroadcastPath...)
	b, _ = WriteVarint(b, uint64(len(f.TrackName)))
	b = append(b, f.TrackName...)
	b, _ = WriteVarint(b, uint64(f.Priority))
	b, _ = WriteVarint(b, f.GroupSequence)
	_, err := w.Write(b)
	return err
}

func (f *FetchMessage) Decode(src io.Reader) error {
	size, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, size)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	str, n, err := ReadString(b)
	if err != nil {
		return err
	}
	f.BroadcastPath = str
	b = b[n:]

	str, n, err = ReadString(b)
	if err != nil {
		return err
	}
	f.TrackName = str
	b = b[n:]

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	f.Priority = uint8(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	f.GroupSequence = num
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
