package json_test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func assertResult(t *testing.T, value map[string]interface{}, expected string) {
	e := make(map[string]interface{})
	err := json.Unmarshal([]byte(expected), &e)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(value, e) {
		bs, _ := json.Marshal(value)
		t.Fatalf("expected:\n%s\n\nactual:\n%s", expected, string(bs))
	}
}
