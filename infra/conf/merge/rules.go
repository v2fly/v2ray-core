package merge

const priorityKey string = "_priority"
const tagKey string = "_tag"

func applyRules(m map[string]interface{}) error {
	err := sortMergeSlices(m)
	if err != nil {
		return err
	}
	removeHelperFields(m)
	return nil
}

// sortMergeSlices enumerates all slices in a map, to sort by priority and merge by tag
func sortMergeSlices(target map[string]interface{}) error {
	for key, value := range target {
		if slice, ok := value.([]interface{}); ok {
			sortByPriority(slice)
			s, err := mergeSameTag(slice)
			if err != nil {
				return err
			}
			target[key] = s
			for _, item := range s {
				if m, ok := item.(map[string]interface{}); ok {
					sortMergeSlices(m)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			sortMergeSlices(field)
		}
	}
	return nil
}

func removeHelperFields(target map[string]interface{}) {
	for key, value := range target {
		if key == priorityKey || key == tagKey {
			delete(target, key)
		} else if slice, ok := value.([]interface{}); ok {
			for _, e := range slice {
				if el, ok := e.(map[string]interface{}); ok {
					removeHelperFields(el)
				}
			}
		} else if field, ok := value.(map[string]interface{}); ok {
			removeHelperFields(field)
		}
	}
}
