package query

import (
	"strings"

	"flowdb/backend/store"
)

func MaskRow(resource string, columns []string, row []any, rules []store.PIIRule) []any {
	if len(rules) == 0 {
		return row
	}
	colIndex := map[string]int{}
	for i, col := range columns {
		colIndex[strings.ToLower(col)] = i
	}
	for _, rule := range rules {
		if !resourceMatch(rule.Resource, resource) {
			continue
		}
		idx, ok := colIndex[strings.ToLower(rule.Field)]
		if !ok || idx >= len(row) {
			continue
		}
		row[idx] = maskValue(row[idx], rule.MaskType)
	}
	return row
}

func MaskDoc(resource string, doc map[string]any, rules []store.PIIRule) map[string]any {
	if len(rules) == 0 {
		return doc
	}
	for _, rule := range rules {
		if !resourceMatch(rule.Resource, resource) {
			continue
		}
		if _, ok := doc[rule.Field]; ok {
			doc[rule.Field] = maskValue(doc[rule.Field], rule.MaskType)
		}
	}
	return doc
}

func resourceMatch(ruleResource, resource string) bool {
	if ruleResource == "*" || strings.EqualFold(ruleResource, resource) {
		return true
	}
	if strings.HasSuffix(ruleResource, "/*") {
		prefix := strings.TrimSuffix(ruleResource, "/*")
		return strings.HasPrefix(resource, prefix)
	}
	return false
}

func maskValue(value any, maskType string) any {
	switch strings.ToLower(maskType) {
	case "null":
		return nil
	default:
		switch value.(type) {
		case string:
			return "****"
		default:
			return "****"
		}
	}
}
