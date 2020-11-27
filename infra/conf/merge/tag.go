package merge

// mergeSameTag do merging-same-tag tasks for all slices in a map
func mergeSameTag(target map[string]interface{}) error {
	for key, value := range target {
		if slice, ok := value.([]interface{}); ok {
			s, err := mergeSameTagSingleSlice(slice)
			if err != nil {
				return err
			}
			target[key] = s
			for _, item := range s {
				if m, ok := item.(map[string]interface{}); ok {
					mergeSameTag(m)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			mergeSameTag(field)
		}
	}
	return nil
}

func mergeSameTagSingleSlice(s []interface{}) ([]interface{}, error) {
	// from: [a,"",b,"",a,"",b,""]
	// to: [a,"",b,"",merged,"",merged,""]
	merged := &struct{}{}
	for i, item1 := range s {
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
				s[j] = merged
				err := mergeMaps(map1, map2)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	ns := make([]interface{}, 0)
	for _, item := range s {
		if item == merged {
			continue
		}
		ns = append(ns, item)
	}
	return ns, nil
}

func getTag(v map[string]interface{}) string {
	if field, ok := v["tag"]; ok {
		if t, ok := field.(string); ok {
			return t
		}
	}
	if field, ok := v[tagKey]; ok {
		if t, ok := field.(string); ok {
			return t
		}
	}
	return ""
}
