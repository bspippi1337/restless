package engine

import (
	"encoding/json"
	"fmt"
)

func PrintPretty(r *Result) {

	out := map[string]interface{}{
		"api_type":  r.APIType,
		"endpoints": r.Endpoints,
		"topology":  r.Topology,
		"workflow":  r.Workflow,
	}

	j, _ := json.MarshalIndent(out, "", "  ")

	fmt.Println(string(j))
}
