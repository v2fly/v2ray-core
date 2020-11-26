package merge

import "sort"

func getPriority(v interface{}) float64 {
	var m map[string]interface{}
	var ok bool
	if m, ok = v.(map[string]interface{}); !ok {
		return 0
	}
	if i, ok := m[priorityKey]; ok {
		if p, ok := i.(float64); ok {
			return p
		}
	}
	return 0
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
