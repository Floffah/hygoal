package datatypes

import (
	"encoding/binary"
	"fmt"
	"io"
)

type HostAddress struct {
	Host string
	Port uint16
}

func DecodeHostAddress(payload []byte, offset int) (*HostAddress, int, error) {
	if offset+2 > len(payload) {
		return nil, 0, io.ErrUnexpectedEOF
	}

	port := binary.LittleEndian.Uint16(payload[offset : offset+2])

	pos := offset + 2

	hostLen, hostLenBytes, err := ReadVarInt(payload, pos)
	if err != nil {
		return nil, 0, fmt.Errorf("host len: %w", err)
	}

	if hostLen < 0 {
		return nil, 0, fmt.Errorf("negative host length")
	}
	if hostLen > 256 {
		return nil, 0, fmt.Errorf("host too long: %d", hostLen)
	}

	start := pos + hostLenBytes
	end := start + hostLen

	if end > len(payload) {
		return nil, 0, io.ErrUnexpectedEOF
	}

	host := string(payload[start:end])

	consumed := end - offset

	return &HostAddress{
		Host: host,
		Port: port,
	}, consumed, nil
}
