package datatypes

import "io"

func ReadVarIntReader(data io.Reader) (value int, size int, _ error) {
	var shift uint = 0
	for {
		var b [1]byte
		_, err := data.Read(b[:])
		if err != nil {
			return 0, size, err
		}
		size++
		value |= int(b[0]&0x7F) << shift
		if (b[0] & 0x80) == 0 {
			break
		}
		shift += 7
		if shift > 28 {
			break
		}
	}
	return value, size, nil
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
