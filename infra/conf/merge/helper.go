package merge

const priorityKey string = "_priority"
const tagKey string = "_tag"

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
