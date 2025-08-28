package rid

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/eknkc/basex"
	"github.com/google/uuid"
)

var (
	ErrInvalidRidFormat = errors.New("RID format is incorrect")
	ErrRidUuidEncoding  = errors.New("RID UUID is not valid base62")
	ErrInvalidRidUuid   = errors.New("the base62 UUID part of the RID string is not a valid UUIDv7")
)

const encodeStr = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" //""
var Base62Encoding, _ = basex.NewEncoding(encodeStr)

type RID struct {
	Prefix string
	ID     uuid.UUID
}

func (R *RID) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Title:       "RID",
		Description: "A type-safe UUID. Prefix indicates Resource type, suffix is a base62 encoded UUIDv7.",
		Format:      "rid",
		Examples:    []any{"u_AZhjuMmhePWkHFALenFEfg"},
	}
}

func (R *RID) UnmarshalJSON(bytes []byte) error {
	var text string
	if err := json.Unmarshal(bytes, &text); err != nil {
		return err
	}

	return R.UnmarshalText([]byte(text))
}

func (R *RID) MarshalJSON() ([]byte, error) {
	return json.Marshal(R.String())
}

func (R *RID) MarshalText() ([]byte, error) {
	return []byte(R.String()), nil
}

func (R *RID) UnmarshalText(text []byte) error {
	textStr := string(text)

	delimiterIdx := strings.IndexRune(textStr, '_')
	if delimiterIdx <= 0 || delimiterIdx >= len(textStr)-1 {
		return ErrInvalidRidFormat
	}

	prefixPart := textStr[:delimiterIdx]
	idPart := textStr[delimiterIdx+1:]

	decodedBytes, decodeErr := Base62Encoding.Decode(idPart)
	if decodeErr != nil {
		return ErrRidUuidEncoding
	}

	decodedUuid := uuid.UUID(decodedBytes)
	if decodedUuid.Version() != 7 {
		return ErrInvalidRidUuid
	}

	R.Prefix = prefixPart
	R.ID = decodedUuid
	return nil
}

func (R *RID) String() string {
	encodedId := Base62Encoding.Encode(R.ID[:])
	// the padding that the base62 library adds isn't really desirable
	encodedId = strings.TrimRight(encodedId, "+")
	return R.Prefix + "_" + encodedId
}

func From(prefix string, id uuid.UUID) RID {
	return RID{
		Prefix: prefix,
		ID:     id,
	}
}

func ParseString(s string) (rid RID, err error) {
	err = rid.UnmarshalText([]byte(s))
	return
}
