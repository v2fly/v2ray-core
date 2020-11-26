package merge

import "sort"

func getPriority(v interface{}) float64 {
	var m map[string]interface{}
	var ok bool
	if m, ok = v.(map[string]interface{}); !ok {
		return 0
	}
	if i, ok := m["_priority"]; ok {
		if p, ok := i.(float64); ok {
			return p
		}
	}
	return 0
}

func getTag(v map[string]interface{}) string {
	if field, ok := v["tag"]; ok {
		if t, ok := field.(string); ok {
			return t
		}
	}
	if field, ok := v["_tag"]; ok {
		if t, ok := field.(string); ok {
			return t
		}
	}
	return ""
}

func mergeSliceItems(s []interface{}) ([]interface{}, error) {
	// from: [a,"",b,"",a,"",b,""]
	// to: [a,"",b,"",nil,"",nil,""]
	for i, item1 := range s {
		// if slice, ok := item.([]interface{}); ok {
		// 	mergeSameTagInSlice(slice)
		// 	continue
		// }
		map1, ok := item1.(map[string]interface{})
		if !ok {
			continue
		}
		tag1 := getTag(map1)
		if tag1 == "" {
			continue
		}
		for j := i + 1; j < len(s); j++ {
			map2, ok := s[j].(map[string]interface{})
			if !ok {
				continue
			}
			tag2 := getTag(map2)
			if tag1 == tag2 {
				s[j] = nil
				m, err := Maps(map1, map2)
				if err != nil {
					return nil, err
				}
				s[i] = m
			}
		}
	}
	ns := make([]interface{}, 0)
	for _, item := range s {
		if item == nil {
			continue
		}
		ns = append(ns, item)
	}
	return ns, nil
}

// sortSlicesInMap sort slices in map by field "priority"
func sortSlicesInMap(target map[string]interface{}) {
	for key, value := range target {
		if slice, ok := value.([]interface{}); ok {
			sort.Slice(slice, func(i, j int) bool { return getPriority(slice[i]) < getPriority(slice[j]) })
			target[key] = slice
		} else if field, ok := value.(map[string]interface{}); ok {
			sortSlicesInMap(field)
		}
	}
}
func removeHelperKey(target map[string]interface{}) {
	for key, value := range target {
		if key == "_priority" || key == "_tag" {
			delete(target, key)
		} else if slice, ok := value.([]interface{}); ok {
			for _, e := range slice {
				if el, ok := e.(map[string]interface{}); ok {
					removeHelperKey(el)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			removeHelperKey(field)
		}
	}
}
func mergeSliceSameTag(target map[string]interface{}) error {
	for key, value := range target {
		if slice, ok := value.([]interface{}); ok {
			s, err := mergeSliceItems(slice)
			if err != nil {
				return err
			}
			target[key] = s
			for _, item := range s {
				if m, ok := item.(map[string]interface{}); ok {
					mergeSliceSameTag(m)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			mergeSliceSameTag(field)
		}
	}
	return nil
}
