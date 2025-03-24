package common

import (
	"encoding/json"
	"strconv"
)

// Int64String supports unmarshaling from both strings and numbers.
type Int64String int64

func (i *Int64String) UnmarshalJSON(data []byte) error {
	// If quoted string
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*i = Int64String(val)
		return nil
	}

	// Else: assume it's a number
	var val int64
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*i = Int64String(val)
	return nil
}
