package sqlite

import (
	"strconv"

	"github.com/bool64/ctxd"
	"github.com/cespare/xxhash/v2"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/usecase/status"
)

type Hash int64

func StringHash(s string) Hash {
	return Hash(xxhash.Sum64String(s))
}

func (h Hash) PrepareJSONSchema(schema *jsonschema.Schema) error {
	*schema.Type = jsonschema.String.Type()

	return nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || data[0] != '"' {
		return status.Wrap(ctxd.SentinelError("string expected"), status.InvalidArgument)
	}

	s := string(data[1 : len(data)-1])

	u, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return err
	}
	*h = Hash(u)

	return nil
}

func (h Hash) String() string {
	res := strconv.FormatUint(uint64(h), 36)

	return res
}

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return status.Wrap(ctxd.SentinelError("non-empty string expected"), status.InvalidArgument)
	}

	u, err := strconv.ParseUint(string(data), 36, 64)
	if err != nil {
		return err
	}
	*h = Hash(u)

	return nil
}
