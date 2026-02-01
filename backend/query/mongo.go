package query

import (
	"encoding/json"
	"strings"
)

type MongoDSL struct {
	Action string `json:"action"`
}

func IsMongoWrite(statement string) bool {
	var dsl MongoDSL
	if err := json.Unmarshal([]byte(statement), &dsl); err != nil {
		return false
	}
	switch strings.ToLower(dsl.Action) {
	case "insert", "update", "delete":
		return true
	default:
		return false
	}
}
