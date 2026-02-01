package policies

import (
	"encoding/json"
	"path"
	"strings"
)

type Document struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Rule struct {
	Effect     string     `json:"effect"`
	Actions    []string   `json:"actions"`
	Resources  []string   `json:"resources"`
	Conditions Conditions `json:"conditions"`
}

type Conditions struct {
	RequireWhere *bool    `json:"require_where,omitempty"`
	MaxRows      *int     `json:"max_rows,omitempty"`
	TimeoutMs    *int     `json:"timeout_ms,omitempty"`
	ReadOnly     *bool    `json:"read_only,omitempty"`
	Environment  []string `json:"environment,omitempty"`
}

type Constraints struct {
	RequireWhere bool
	MaxRows      int
	TimeoutMs    int
	ReadOnly     bool
}

type Engine struct {
	policies []Document
}

func NewEngine(rawPolicies [][]byte) (*Engine, error) {
	var docs []Document
	for _, raw := range rawPolicies {
		var doc Document
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return &Engine{policies: docs}, nil
}

func (e *Engine) Evaluate(action string, resource string, env string) (bool, Constraints) {
	allowed := false
	constraints := Constraints{
		MaxRows:   0,
		TimeoutMs: 0,
	}
	for _, doc := range e.policies {
		for _, rule := range doc.Rules {
			if strings.ToLower(rule.Effect) != "allow" {
				continue
			}
			if !matchesAny(rule.Actions, action) {
				continue
			}
			if !matchesAnyResource(rule.Resources, resource) {
				continue
			}
			if len(rule.Conditions.Environment) > 0 && !matchesAny(rule.Conditions.Environment, env) {
				continue
			}
			allowed = true
			if rule.Conditions.RequireWhere != nil && *rule.Conditions.RequireWhere {
				constraints.RequireWhere = true
			}
			if rule.Conditions.ReadOnly != nil && *rule.Conditions.ReadOnly {
				constraints.ReadOnly = true
			}
			if rule.Conditions.MaxRows != nil {
				if constraints.MaxRows == 0 || *rule.Conditions.MaxRows < constraints.MaxRows {
					constraints.MaxRows = *rule.Conditions.MaxRows
				}
			}
			if rule.Conditions.TimeoutMs != nil {
				if constraints.TimeoutMs == 0 || *rule.Conditions.TimeoutMs < constraints.TimeoutMs {
					constraints.TimeoutMs = *rule.Conditions.TimeoutMs
				}
			}
		}
	}
	return allowed, constraints
}

func (e *Engine) HasRules() bool {
	if e == nil {
		return false
	}
	for _, doc := range e.policies {
		if len(doc.Rules) > 0 {
			return true
		}
	}
	return false
}

func matchesAny(list []string, value string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func matchesAnyResource(patterns []string, resource string) bool {
	for _, pat := range patterns {
		if pat == "*" {
			return true
		}
		if pathMatch(pat, resource) {
			return true
		}
	}
	return false
}

func pathMatch(pattern, resource string) bool {
	pattern = strings.ReplaceAll(pattern, "**", "*")
	match, err := path.Match(pattern, resource)
	if err != nil {
		return false
	}
	return match
}
