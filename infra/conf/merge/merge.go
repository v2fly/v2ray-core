// Copyright 2020 Jebbs. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

/*
Package merge provides the capability to merge multiple
JSON files or contents into one output.

Merge Rules:

- Simple values (string, number, boolean) are overwritten, others are merged
- Elements with same "tag" (or "_tag") in an array will be merged
- Add "_priority" property to array elements will help sort the
*/
package merge

import (
	"bytes"
	"encoding/json"

	"github.com/v2fly/v2ray-core/v5/infra/conf/serial"
)

// JSONs merges multiple json contents into one json.
func JSONs(args [][]byte) ([]byte, error) {
	m := make(map[string]interface{})
	for _, arg := range args {
		if _, err := ToMap(arg, m); err != nil {
			return nil, err
		}
	}
	return FromMap(m)
}

// ToMap merges json content to target map and returns it
func ToMap(content []byte, target map[string]interface{}) (map[string]interface{}, error) {
	if target == nil {
		target = make(map[string]interface{})
	}
	r := bytes.NewReader(content)
	n := make(map[string]interface{})
	err := serial.DecodeJSON(r, &n)
	if err != nil {
		return nil, err
	}
	if err = mergeMaps(target, n); err != nil {
		return nil, err
	}
	return target, nil
}

// FromMap apply merge rules to map and convert it to json
func FromMap(target map[string]interface{}) ([]byte, error) {
	if target == nil {
		target = make(map[string]interface{})
	}
	err := ApplyRules(target)
	if err != nil {
		return nil, err
	}
	return json.Marshal(target)
}
