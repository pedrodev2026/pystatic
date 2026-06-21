package runtime

import (
	"encoding/json"
)

func JsonEncode(v any) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func JsonDecode(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
