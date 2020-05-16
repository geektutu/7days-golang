package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const MagicNumber int32 = 0xECABCD

type SerializeType int8

const (
	JSON SerializeType = iota
)

var Codecs = map[SerializeType]Codec{
	JSON: &JSONCodec{},
}

type Status int8

const (
	OK Status = iota
	ExecError
	NotFoundError
)

type Header struct {
	Magic             int32
	Status            Status
	SerializeType     SerializeType
	ServiceMethodSize int32
	PayloadSize       int32
}

type Message struct {
	*Header
	ServiceMethod string
	Payload       []byte
}

func NewMessage() *Message {
	return &Message{
		Header: &Header{Magic: MagicNumber},
	}
}

func (m *Message) HandleError(status Status, err error) *Message {
	m.Status = status
	_ = m.SetPayload(err)
	return m
}

func (m *Message) SetServiceMethod(name string) {
	m.ServiceMethod = name
}

func (m *Message) GetServiceMethod() (service, method string, err error) {
	parts := strings.Split(m.ServiceMethod, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("<service.method> format error")
	}
	return parts[0], parts[1], nil
}

func (m *Message) GetPayload(i interface{}) error {
	return Codecs[m.SerializeType].Decode(m.Payload, i)
}

func (m *Message) SetPayload(i interface{}) (err error) {
	m.Payload, err = Codecs[m.SerializeType].Encode(i)
	return
}

func (m *Message) Clone() *Message {
	m2 := NewMessage()
	*m2.Header = *m.Header
	m2.ServiceMethod = m.ServiceMethod
	return m2
}
func Read(r io.Reader) (*Message, error) {
	m := NewMessage()
	if err := binary.Read(r, binary.BigEndian, m.Header); err != nil {
		return nil, err
	}
	if m.Magic != MagicNumber {
		return nil, fmt.Errorf("invalid message: wrong magic number")
	}

	buf := make([]byte, m.ServiceMethodSize+m.PayloadSize)
	if err := binary.Read(r, binary.BigEndian, buf); err != nil {
		return nil, err
	}
	m.ServiceMethod = string(buf[:m.ServiceMethodSize])
	m.Payload = buf[m.ServiceMethodSize:]
	return m, nil
}
func (m *Message) Write(w io.Writer) error {
	m.PayloadSize = int32(len(m.Payload))
	m.ServiceMethodSize = int32(len(m.ServiceMethod))
	buf := bytes.NewBufferString(m.ServiceMethod)
	buf.Write(m.Payload)
	if err := binary.Write(w, binary.BigEndian, m.Header); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, buf.Bytes()); err != nil {
		return err
	}
	return nil
}
