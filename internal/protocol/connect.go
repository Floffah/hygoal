package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hygoal/internal/datatypes"
)

type ClientType byte

const (
	ClientTypeGame   ClientType = 0 // confirm from Java enum values
	ClientTypeEditor ClientType = 1
)

type Connect struct {
	ProtocolHash   string
	ClientType     ClientType
	Language       *string
	IdentityToken  *string
	UUID           [16]byte // or uuid.UUID if you like
	Username       string
	ReferralData   []byte
	ReferralSource *datatypes.HostAddress
}

func (*Connect) ID() uint32 { return 0 }

const varStart = 102

func DecodeConnect(payload []byte) (Packet, error) {
	if len(payload) < 102 {
		return nil, fmt.Errorf("connect payload too small: %d", len(payload))
	}

	// optional fields bitfield
	var nullBits byte = payload[0]

	// fixed fields
	protoRaw := payload[1:65] // 64 bytes
	protocolHash := string(bytes.TrimRight(protoRaw, "\x00"))

	clientType := payload[65]

	var uuid [16]byte
	copy(uuid[:], payload[66:82])

	// offsets (signed int32 LE)
	langOffset := int(int32(binary.LittleEndian.Uint32(payload[82:86])))
	tokenOffset := int(int32(binary.LittleEndian.Uint32(payload[86:90])))
	userOffset := int(int32(binary.LittleEndian.Uint32(payload[90:94])))
	refDataOffset := int(int32(binary.LittleEndian.Uint32(payload[94:98])))
	refSrcOffset := int(int32(binary.LittleEndian.Uint32(payload[98:102])))

	// read username
	usernamePos := varStart + userOffset
	username, _, err := datatypes.ReadVarString(payload, usernamePos, 16, true)
	if err != nil {
		return nil, fmt.Errorf("error reading username: %v", err)
	}

	out := &Connect{
		ProtocolHash: protocolHash,
		ClientType:   ClientType(clientType),
		UUID:         uuid,
		Username:     username,
	}

	// read optional language
	if (nullBits & 0x01) != 0 {
		langPos := varStart + langOffset
		language, _, err := datatypes.ReadVarString(payload, langPos, 128, true)
		if err != nil {
			return nil, fmt.Errorf("error reading language: %v", err)
		}
		out.Language = &language
	}

	// read optional identity token
	if (nullBits & 0x02) != 0 {
		tokenPos := varStart + tokenOffset
		identityToken, _, err := datatypes.ReadVarString(payload, tokenPos, 8192, false)
		if err != nil {
			return nil, fmt.Errorf("error reading identity token: %v", err)
		}
		out.IdentityToken = &identityToken
	}

	// read referral data
	if (nullBits & 0x04) != 0 {
		refDataPos := varStart + refDataOffset
		refDataLen, refDataLenSize, err := datatypes.ReadVarInt(payload, refDataPos)
		if err != nil {
			return nil, fmt.Errorf("error reading referral data length: %v", err)
		}
		if refDataLen < 0 || refDataLen > 4096 {
			return nil, fmt.Errorf("referralData len %d", refDataLen)
		}

		start := refDataPos + refDataLenSize
		end := start + refDataLen
		if end > len(payload) {
			return nil, fmt.Errorf("referralData exceeds payload length")
		}

		out.ReferralData = make([]byte, refDataLen)
		copy(out.ReferralData, payload[start:end])
	}

	// read host address
	if (nullBits & 0x08) != 0 {
		refSrcPos := varStart + refSrcOffset
		referralSource, _, err := datatypes.DecodeHostAddress(payload, refSrcPos)
		if err != nil {
			return nil, fmt.Errorf("error reading referral source: %v", err)
		}
		out.ReferralSource = referralSource
	}

	return out, nil
}
