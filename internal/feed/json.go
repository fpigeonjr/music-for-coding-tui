package feed

import "encoding/json"

// unmarshalString decodes a JSON string literal (e.g. `"hello\nworld"`) into s.
func unmarshalString(data []byte, s *string) error {
	return json.Unmarshal(data, s)
}
