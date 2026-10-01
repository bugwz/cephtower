package mutation

import (
	"bytes"
	"encoding/json"
	"io"
)

func crushRuleCreated(parameters map[string]any, data []byte) bool {
	var rule struct {
		Name  string `json:"rule_name"`
		Type  int    `json:"type"`
		Steps []struct {
			Op       string `json:"op"`
			ItemName string `json:"item_name"`
			Type     string `json:"type"`
			Num      *int   `json:"num"`
		} `json:"steps"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&rule) != nil || decoder.Decode(new(any)) != io.EOF || rule.Name != optional(parameters, "name") || rule.Type != 1 || len(rule.Steps) != 3 {
		return false
	}
	root := optional(parameters, "root")
	if class := optional(parameters, "device_class"); class != "" {
		root += "~" + class
	}
	failure := optional(parameters, "failure_domain")
	if failure == "" {
		failure = "host"
	}
	op := "chooseleaf_firstn"
	if failure == "osd" {
		op = "choose_firstn"
	}
	return rule.Steps[0].Op == "take" && rule.Steps[0].ItemName == root && rule.Steps[1].Op == op && rule.Steps[1].Type == failure && rule.Steps[1].Num != nil && *rule.Steps[1].Num == 0 && rule.Steps[2].Op == "emit"
}
