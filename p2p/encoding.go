package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, interface{}) error
}

type GOBDecoder struct{}

func (d GOBDecoder) Decode(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}
