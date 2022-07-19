package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func JsonDataToBuff(payload map[string]interface{}) *bytes.Buffer {
	pbytes, _ := json.Marshal(payload)
	buff := bytes.NewBuffer(pbytes)
	return buff
}

// https://stackoverflow.com/questions/21197239/decoding-json-using-json-unmarshal-vs-json-newdecoder-decode
func GetHttpJson(r http.Response, target interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
