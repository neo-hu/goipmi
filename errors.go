// +build linux

package goipmi

import (
	"encoding/hex"
	"fmt"
)

type UnsupportedSDRTypeErr struct {
	sdrType byte
	nextId  uint16
}

func (c UnsupportedSDRTypeErr) Error() string {
	return fmt.Sprintf("Unsupported SDR type(0x%s)", hex.EncodeToString([]byte{c.sdrType}))
}

func NewUnsupportedSDRTypeErr(sdrType byte, nextId uint16) UnsupportedSDRTypeErr {
	return UnsupportedSDRTypeErr{sdrType: sdrType, nextId: nextId}
}

func IsUnsupportedSDRTypeErr(err error) bool {
	switch err.(type) {
	case UnsupportedSDRTypeErr:
		return true
	}
	return false
}
