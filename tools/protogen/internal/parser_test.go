package protogen

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestBasicPacket(t *testing.T) {
	parser := NewParser(`
	packet 1 LoginRequest {
		username string
		password string
		@someFixedField int32
        @someOptionalField? int64
		@someBitSizeField string[12]
	}
	`)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatal(FormatParseError(err, "unknown"))
	}

	snaps.MatchSnapshot(t, ast)
}

func TestBasicType(t *testing.T) {
	parser := NewParser(`
	type HostAddress {
		port uint16
		hostname string
	}
	`)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatal(FormatParseError(err, "unknown"))
	}

	snaps.MatchSnapshot(t, ast)
}
