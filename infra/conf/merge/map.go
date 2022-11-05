// Copyright 2020 Jebbs. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package merge

import (
	"fmt"
	"reflect"
)

// mergeMaps merges source map into target
// it supports only map[string]interface{} type for any children of the map tree
func mergeMaps(target map[string]interface{}, source map[string]interface{}) (err error) {
	for key, value := range source {
		target[key], err = mergeField(target[key], value)
		if err != nil {
			return
		}
	}
	return
}

func mergeField(target interface{}, source interface{}) (interface{}, error) {
	if source == nil {
		return target, nil
	}
	if target == nil {
		return source, nil
	}
	if reflect.TypeOf(source) != reflect.TypeOf(target) {
		return nil, fmt.Errorf("type mismatch, expect %T, incoming %T", target, source)
	}
	if slice, ok := source.([]interface{}); ok {
		tslice, _ := target.([]interface{})
		tslice = append(tslice, slice...)
		return tslice, nil
	} else if smap, ok := source.(map[string]interface{}); ok {
		tmap, _ := target.(map[string]interface{})
		err := mergeMaps(tmap, smap)
		return tmap, err
	}
	return source, nil
}
