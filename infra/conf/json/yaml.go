package json

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// FromYAML convert yaml to json
func FromYAML(v []byte) ([]byte, error) {
	m1 := make(map[interface{}]interface{})
	err := yaml.Unmarshal(v, &m1)
	if err != nil {
		return nil, err
	}
	m2 := convert(m1)
	j, err := json.Marshal(m2)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func convert(m map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		var value interface{}
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			value = convert(v2)
		case []interface{}:
			for i, el := range v2 {
				if m, ok := el.(map[interface{}]interface{}); ok {
					v2[i] = convert(m)
				}
			}
			value = v2
		default:
			value = v
		}
		key := "null"
		if k != nil {
			key = fmt.Sprint(k)
		}
		res[key] = value
	}
	return res
}
