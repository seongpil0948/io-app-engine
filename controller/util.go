package controller

import "encoding/json"

func dump(x interface{}) string {
	json, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(json)
}
