package datatypes

import (
	"fmt"
	"io"
)

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
