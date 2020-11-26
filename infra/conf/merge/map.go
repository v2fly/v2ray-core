package merge

import (
	"fmt"
	"io"
	"reflect"

	"v2ray.com/core/infra/conf/serial"
)

// Maps merges source map into target, and return it
func Maps(target map[string]interface{}, source map[string]interface{}) (out map[string]interface{}, err error) {
	for key, value := range source {
		// fmt.Printf("[%s] type: %s, kind: %s\n", key, getType(fieldTypeSrc.Type).Name(), getType(fieldTypeSrc.Type).Kind())
		target[key], err = mergeField(target[key], value)
		if err != nil {
			return nil, err
		}
	}
	return target, nil
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
			target = append(tslice, slice...)
			return target, nil
		}
		return nil, fmt.Errorf("value type mismatch, source is 'slice' but target not: %s", source)
	} else if smap, ok := source.(map[string]interface{}); ok {
		if tmap, ok := target.(map[string]interface{}); ok {
			_, err := Maps(tmap, smap)
			return tmap, err
		}
		return nil, fmt.Errorf("value type mismatch, source is 'map[string]interface{}' but target not: %s", source)
	}
	return source, nil
}

func jsonToMap(r io.Reader) (map[string]interface{}, error) {
	c := make(map[string]interface{})
	err := serial.DecodeJSON(r, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
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
