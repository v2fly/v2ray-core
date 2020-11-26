package merge_test

import (
	"testing"

	"v2ray.com/core/infra/conf/merge"
)

func TestMergeJSON(t *testing.T) {

	json1 := `
	{
	  "log": {"access": "some_value", "loglevel": "debug"},
	  "inbounds": [{"tag": "in-1"}],
	  "outbounds": [{"priority": 100, "tag": "out-1"}],
	  "routing": {"rules": [{"inboundTag":["in-1"],"outboundTag":"out-1"}]}
	}
`
	json2 := `
	{
	  "log": {"loglevel": "error"},
	  "inbounds": [{"tag": "in-2"}],
	  "outbounds": [{"priority": -100, "tag": "out-2"}],
	  "routing": {"rules": [{"inboundTag":["in-2"],"outboundTag":"out-2"}]}
	}	
`
	m, err := merge.JSONsToMap([][]byte{[]byte(json1), []byte(json2)})
	if err != nil {
		t.Error(err)
	}
	assertField(t, m, []interface{}{"log", "access"}, "some_value")
	assertField(t, m, []interface{}{"log", "loglevel"}, "error")
	assertField(t, m, []interface{}{"outbounds", 0, "tag"}, "out-2")
	assertField(t, m, []interface{}{"outbounds", 1, "tag"}, "out-1")
	assertField(t, m, []interface{}{"routing", "rules", 1, "inboundTag", 0}, "in-2")
}

func assertField(t *testing.T, m map[string]interface{}, path []interface{}, value string) {
	var cur interface{}
	cur = m
	for i, key := range path {
		if k, ok := key.(string); ok {
			c, ok := cur.(map[string]interface{})
			if !ok {
				t.Fatalf("no field for %s: %v[%d]", k, path, i)
			}
			cur, ok = c[k]
			if !ok {
				t.Fatalf("%s not found: %s", k, path)
			}
			continue
		}
		if k, ok := key.(int); ok {
			c, ok := cur.([]interface{})
			if !ok {
				t.Fatalf("not a slice for %d: %v[%d]", k, path, i)
			}
			if k < 0 || k > len(c)-1 {
				t.Fatalf("%d out of range for %v[%d]: %v", k, path, i, c)
			}
			cur = c[k]
			continue
		}
	}
	v, ok := cur.(string)
	if !ok || v != value {
		t.Fatalf("%v: value mismatch, expected: %s, actual: %s", path, value, v)
	}
}
