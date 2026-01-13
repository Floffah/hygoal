package protocol

import (
	"fmt"
)

type Packet interface {
	ID() uint32
}

type Decoder func(payload []byte) (Packet, error)

var decoders = map[uint32]Decoder{}

func Register(id uint32, d Decoder) {
	if _, exists := decoders[id]; exists {
		panic("duplicate packet id")
	}
	decoders[id] = d
}

func DecodeByID(id uint32, payload []byte) (Packet, error) {
	d, ok := decoders[id]
	if !ok {
		return nil, fmt.Errorf("unknown packet id %d", id)
	}
	return d(payload)
}

func init() {
	Register(0, DecodeConnect)
}
