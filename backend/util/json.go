package util

import (
	"bytes"
	"encoding/json"
	"sort"
)

type orderedMap []orderedKV

type orderedKV struct {
	Key string
	Val any
}

func (o orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, kv := range o {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func canonicalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(orderedMap, 0, len(keys))
		for _, k := range keys {
			out = append(out, orderedKV{Key: k, Val: canonicalize(t[k])})
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, v := range t {
			out = append(out, canonicalize(v))
		}
		return out
	default:
		return v
	}
}

func CanonicalJSON(v any) ([]byte, error) {
	return json.Marshal(canonicalize(v))
}
