package data_type

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FlexibleInt64 int64

func (v FlexibleInt64) Int64() int64 {
	return int64(v)
}

func (v FlexibleInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Int64())
}

func (v *FlexibleInt64) UnmarshalJSON(b []byte) error {
	raw := strings.TrimSpace(string(b))
	if raw == "" || raw == "null" {
		*v = 0
		return nil
	}

	var n int64
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		parsed, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 string %q: %w", s, err)
		}
		n = parsed
	} else {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 value %q: %w", raw, err)
		}
		n = parsed
	}

	*v = FlexibleInt64(n)
	return nil
}
