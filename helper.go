// +build linux

package goipmi

import (
	"github.com/pkg/errors"
)

var (
	DataTooShort = errors.New("Data too short for message")
)

type ByteBuffer struct {
	b []byte
}

func NewByteBuffer(b []byte) *ByteBuffer {
	return &ByteBuffer{
		b: b,
	}
}

func (b *ByteBuffer) checkLen(length int) error {
	if len(b.b) < length {
		return DataTooShort
	}
	return nil
}

func (b *ByteBuffer) PopString(length int) (val string, err error) {
	if err := b.checkLen(length); err != nil {
		return "", err
	}
	val = string(b.b[:length])
	b.b = b.b[length:]
	return
}
func (b *ByteBuffer) PopSlice(length int) (*ByteBuffer, error) {
	if err := b.checkLen(length); err != nil {
		return nil, err
	}
	data := NewByteBuffer(b.b[:length])
	b.b = b.b[length:]
	return data, nil
}

func (b *ByteBuffer) Len() int {
	return len(b.b)
}

func (b *ByteBuffer) PopUint8() (val uint8, err error) {
	if err = b.checkLen(1); err != nil {
		return 0, err
	}
	val = b.b[0]
	b.b = b.b[1:]
	return
}

func (b *ByteBuffer) PopUint16() (val uint16, err error) {
	if err = b.checkLen(2); err != nil {
		return 0, err
	}
	val = uint16(b.b[0]) | uint16(b.b[1])<<8
	b.b = b.b[2:]
	return
}

func (b *ByteBuffer) PopUint24() (val uint32, err error) {
	if err = b.checkLen(3); err != nil {
		return 0, err
	}
	val = uint32(b.b[0]) | uint32(b.b[1])<<8 | uint32(b.b[2])<<16
	b.b = b.b[3:]
	return
}

func (b *ByteBuffer) PopUint32() (val uint32, err error) {
	if err = b.checkLen(4); err != nil {
		return 0, err
	}
	val = uint32(b.b[0]) | uint32(b.b[1])<<8 | uint32(b.b[2])<<16 | uint32(b.b[3])<<24
	b.b = b.b[4:]
	return
}
