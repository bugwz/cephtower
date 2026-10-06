package ceph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type erasureProfileWire map[string]string

func (profile *erasureProfileWire) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("erasure profile must be an object")
	}
	values := make(erasureProfileWire)
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("invalid erasure profile parameter name")
		}
		if _, exists := values[key]; exists {
			return fmt.Errorf("duplicate erasure profile parameter")
		}
		var value *string
		if err := decoder.Decode(&value); err != nil || value == nil {
			return fmt.Errorf("erasure profile parameters must be strings")
		}
		values[key] = *value
	}
	token, err = decoder.Token()
	if err != nil || token != json.Delim('}') || decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("invalid erasure profile object")
	}
	*profile = values
	return nil
}
