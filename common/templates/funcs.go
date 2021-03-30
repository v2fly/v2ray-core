package templates

import "text/template"

var AssistFunctions template.FuncMap

func RegisterFunction(name string, function interface{}) {
	if AssistFunctions == nil {
		AssistFunctions = map[string]interface{}{}
	}
	AssistFunctions[name] = function
}

func Dec(val int) int {
	return val - 1
}

func ShortHand(val string) string {
	return val[:6]
}

func init() {
	RegisterFunction("dec", Dec)
	RegisterFunction("shorthand", ShortHand)
}
