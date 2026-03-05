package probe

import "encoding/json"

func SimpleJSONBody() []byte {
	m := map[string]any{
		"name": "restless",
		"id":   1,
		"flag": true,
	}

	b, _ := json.Marshal(m)
	return b
}
