package merge

import (
	"fmt"
	"reflect"
)

// mergeMaps merges source map into target
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
	if (source == nil) || isZero(source) {
		return target, nil
	}
	if target == nil || isZero(target) {
		return source, nil
	}
	if slice, ok := source.([]interface{}); ok {
		if tslice, ok := target.([]interface{}); ok {
			tslice = append(tslice, slice...)
			return tslice, nil
		}
		return nil, fmt.Errorf("value type mismatch, source is 'slice' but target not: %s", source)
	} else if smap, ok := source.(map[string]interface{}); ok {
		if tmap, ok := target.(map[string]interface{}); ok {
			err := mergeMaps(tmap, smap)
			return tmap, err
		}
		return nil, fmt.Errorf("value type mismatch, source is 'map[string]interface{}' but target not: %s", source)
	}
	return source, nil
}

func isZero(v interface{}) bool {
	return getValue(reflect.ValueOf(v)).IsZero()
}

func getValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}
