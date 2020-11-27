package merge_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"v2ray.com/core/infra/conf/merge"
	"v2ray.com/core/infra/conf/serial"
)

func TestMergeJSON(t *testing.T) {
	json1 := `
	  {
		"log": {"access": "some_value", "loglevel": "debug"},
		"inbounds": [{"tag": "in-1"}],
		"outbounds": [{"_priority": 100, "tag": "out-1"}],
		"routing": {"rules": [
		  {"_tag":"default_route","inboundTag":["in-1"],"outboundTag":"out-1"}
		]}
	  }
`
	json2 := `
	  {
		"log": {"loglevel": "error"},
		"inbounds": [{"tag": "in-2"}],
		"outbounds": [{"_priority": -100, "tag": "out-2"}],
		"routing": {"rules": [
		  {"inboundTag":["in-2"],"outboundTag":"out-2"},
		  {"_tag":"default_route","inboundTag":["in-1.1"],"outboundTag":"out-1.1"}
		]}
	  }
`
	expected := `
	{
	  "log": {"access": "some_value", "loglevel": "error"},
	  "inbounds": [{"tag": "in-1"},{"tag": "in-2"}],
	  "outbounds": [
		   {"tag": "out-2"},
		   {"tag": "out-1"}
	  ],
	  "routing": {"rules": [
		   {"inboundTag":["in-1","in-1.1"],"outboundTag":"out-1.1"},
		   {"inboundTag":["in-2"],"outboundTag":"out-2"}
	  ]}
	}
	`
	m, err := merge.BytesToMap([][]byte{[]byte(json1), []byte(json2)})
	if err != nil {
		t.Error(err)
	}
	assertResult(t, m, expected)
}

func TestMergeJSON_MergeTag(t *testing.T) {
	json1 := `
	{
	  	"routing": {
		  	"rules": [{
				"tag":"1",
				"inboundTag": ["in-1"],
				"outboundTag": "out-1"
			}]
		}
	}
`
	json2 := `
	{
	  	"routing": {
		  	"rules": [{
				"_tag":"1",
				"inboundTag": ["in-2"],
				"outboundTag": "out-2"
			}]
		}
	}	
`
	expected := `
	{
	  	"routing": {
			"rules": [{
				"tag":"1",
				"inboundTag": ["in-1", "in-2"],
				"outboundTag": "out-2"
			}]
	  	}
	}
	`
	m, err := merge.BytesToMap([][]byte{[]byte(json1), []byte(json2)})
	if err != nil {
		t.Error(err)
	}
	assertResult(t, m, expected)
}

func TestMergeJSON_MergeTag2(t *testing.T) {
	json1 := `
	{
	  	"array_1": [{
			"_tag":"1",
			"array_2": [{
				"_tag":"2",
				"array_3.1": ["string",true,false],
				"array_3.2": [1,2,3],
				"number_1": 1,
				"number_2": 1,
				"bool_1": true,
				"bool_2": true
			}]
		}]
	}
`
	json2 := `
	{
		"array_1": [{
			"_tag":"1",
			"array_2": [{
				"_tag":"2",
				"array_3.1": [0,1,null],
				"array_3.2": null,
				"number_1": 0,
				"number_2": 1,
				"bool_1": true,
				"bool_2": false,
				"null_1": null
			}]
		}]
	}
`
	expected := `
	{
	  "array_1": [{
		"array_2": [{
			"array_3.1": ["string",true,false,0,1,null],
			"array_3.2": [1,2,3],
			"number_1": 0,
			"number_2": 1,
			"bool_1": true,
			"bool_2": false,
			"null_1": null
		}]
	  }]
	}
	`
	m, err := merge.BytesToMap([][]byte{[]byte(json1), []byte(json2)})
	if err != nil {
		t.Error(err)
	}
	assertResult(t, m, expected)
}

func assertResult(t *testing.T, value map[string]interface{}, expected string) {
	e := make(map[string]interface{})
	err := serial.DecodeJSON(strings.NewReader(expected), &e)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(value, e) {
		bs, _ := json.Marshal(value)
		t.Fatalf("expected:\n%s\n\nactual:\n%s", expected, string(bs))
	}
}
