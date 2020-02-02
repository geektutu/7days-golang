package geecache

import "errors"

type Sink interface {
	SetBytes(v []byte) error
	view() (ByteView, error)
}

type allocBytesSink struct {
	dst *[]byte
	v   ByteView
}

func AllocatingByteSliceSink(dst *[]byte) Sink {
	return &allocBytesSink{dst: dst}
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (s *allocBytesSink) SetBytes(b []byte) error {
	*s.dst = cloneBytes(b)
	s.v.b = b
	return nil
}

func (s *allocBytesSink) view() (ByteView, error) {
	if s.v.b == nil {
		return ByteView{}, errors.New("byteview not set")
	}
	return s.v, nil
}
