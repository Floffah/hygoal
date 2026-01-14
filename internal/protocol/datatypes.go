package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

func DecodeHostAddress(payload []byte, offset int) (HostAddress, int, error) {
	if offset+2 > len(payload) {
		return HostAddress{}, 0, io.ErrUnexpectedEOF
	}

	port := binary.LittleEndian.Uint16(payload[offset : offset+2])

	pos := offset + 2

	hostLen, hostLenBytes, err := ReadVarInt(payload, pos)
	if err != nil {
		return HostAddress{}, 0, fmt.Errorf("host len: %w", err)
	}

	if hostLen < 0 {
		return HostAddress{}, 0, fmt.Errorf("negative host length")
	}
	if hostLen > 256 {
		return HostAddress{}, 0, fmt.Errorf("host too long: %d", hostLen)
	}

	start := pos + hostLenBytes
	end := start + hostLen

	if end > len(payload) {
		return HostAddress{}, 0, io.ErrUnexpectedEOF
	}

	host := string(payload[start:end])

	consumed := end - offset

	return HostAddress{
		Hostname: host,
		Port:     port,
	}, consumed, nil
}

func ReadVarInt(data []byte, pos int) (value int, size int, _ error) {
	var shift uint = 0
	for {
		if pos+size >= len(data) {
			return 0, size, io.ErrUnexpectedEOF
		}
		b := data[pos+size]
		size++
		value |= int(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift > 28 {
			break
		}
	}
	return value, size, nil
}

func ReadVarString(payload []byte, pos int, max int, ascii bool) (string, int, error) {
	n, nLen, err := ReadVarInt(payload, pos)
	if err != nil {
		return "", 0, err
	}
	if n < 0 || n > max {
		return "", 0, fmt.Errorf("var string len %d > max %d", n, max)
	}
	start := pos + nLen
	end := start + n
	if end > len(payload) {
		return "", 0, io.ErrUnexpectedEOF
	}
	b := payload[start:end]
	if ascii {
		// Java uses PacketIO.ASCII for language & username
		// likely means bytes->string without utf8 validation
		return string(b), nLen + n, nil
	}
	// identity token is UTF-8
	return string(b), nLen + n, nil
}
