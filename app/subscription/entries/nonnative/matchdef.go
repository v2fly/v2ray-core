package nonnative

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"strings"
	"text/template"
)

//go:embed definitions/*
var embeddedDefinitions embed.FS

func NewDefMatcher() *DefMatcher {
	d := &DefMatcher{}
	d.init()
	return d
}

type DefMatcher struct {
	templates *template.Template
}

type ExecutionEnvironment struct {
	link AbstractNonNativeLink
}

func (d *DefMatcher) createFuncMap() template.FuncMap {
	return map[string]any{
		"assertExists": func(env *ExecutionEnvironment, names ...string) (bool, error) {
			link := env.link
			for _, v := range names {
				_, ok := link.Values[v]
				if !ok {
					return false, newError("failed assertExists of ", v)
				}
			}
			return true, nil
		},
		"assertIsOneOf": func(env *ExecutionEnvironment, name string, values ...string) (bool, error) {
			link := env.link
			actualValue, ok := link.Values[name]
			if !ok {
				return false, newError("failed assertIs of non-exist ", name)
			}
			found := false
			for _, currentValue := range values {
				if currentValue == actualValue {
					found = true
					break
				}
			}
			if !found {
				return false, newError("failed assertIsOneOf name = ", actualValue, "is not one of ", values)
			}
			return true, nil
		},
		"assertValueIsOneOf": func(value string, values ...string) (bool, error) {
			actualValue := value
			found := false
			for _, currentValue := range values {
				if currentValue == actualValue {
					found = true
					break
				}
			}
			if !found {
				return false, newError("failed assertIsOneOf name = ", actualValue, "is not one of ", values)
			}
			return true, nil
		},
		"tryGet": func(env *ExecutionEnvironment, names ...string) (string, error) {
			link := env.link
			for _, currentName := range names {
				value, ok := link.Values[currentName]
				if ok {
					return value, nil
				} else if currentName == "<default>" {
					return "", nil
				}
			}
			return "", newError("failed tryGet exists none of ", names)
		},
		"splitAndGetNth": func(sep string, n int, content string) (string, error) {
			result := strings.Split(content, sep)
			if n > len(result)-1 {
				return "", newError("failed splitAndGetNth exists too short content:", content, "n = ", n, "sep =", sep)
			}
			if n < 0 {
				n = len(result) + n
				if n < 0 {
					return "", newError("failed splitAndGetNth exists too short content:", content, "n = ", n, "sep =", sep)
				}
			}
			return result[n], nil
		},
		"splitAndGetAfterNth": func(sep string, n int, content string) ([]string, error) {
			result := strings.Split(content, sep)
			if n < 0 {
				n = len(result) + n
			}
			if n > len(result)-1 {
				return []string{}, newError("failed splitAndGetNth exists too short content:", content)
			}
			return result[n:], nil
		},
		"splitAndGetBeforeNth": func(sep string, n int, content string) ([]string, error) {
			result := strings.Split(content, sep)
			if n < 0 {
				n = len(result) + n
			}
			if n > len(result)-1 {
				return []string{}, newError("failed splitAndGetNth exists too short content:", content)
			}
			return result[:n], nil
		},
		"jsonEncode": func(content any) (string, error) {
			buf := bytes.NewBuffer(nil)
			err := json.NewEncoder(buf).Encode(content)
			if err != nil {
				return "", newError("unable to jsonQuote ", content).Base(err)
			}
			return buf.String(), nil
		},
		"stringCutSuffix": func(suffix, content string) (string, error) {
			remaining, found := strings.CutSuffix(content, suffix)
			if !found {
				return "", newError("suffix not found in content =", suffix, " suffix =", suffix)
			}
			return remaining, nil
		},
		"unalias": func(standardName string, names ...string) (string, error) {
			if len(names) == 0 {
				return "", newError("no input value specified")
			}
			actualInput := names[len(names)-1]
			alias := names[:len(names)-1]
			for _, v := range alias {
				if v == actualInput {
					return standardName, nil
				}
			}
			return actualInput, nil
		},
	}
}

func (d *DefMatcher) init() {
	d.templates = template.New("root").Funcs(d.createFuncMap())
}

func (d *DefMatcher) LoadEmbeddedDefinitions() error {
	return d.LoadDefinitions(embeddedDefinitions)
}

func (d *DefMatcher) LoadDefinitions(fs fs.FS) error {
	var err error
	d.templates, err = d.templates.ParseFS(fs, "definitions/*.jsont")
	if err != nil {
		return err
	}
	return nil
}

func (d *DefMatcher) ExecuteNamed(link AbstractNonNativeLink, name string) ([]byte, error) {
	outputBuffer := bytes.NewBuffer(nil)
	env := &ExecutionEnvironment{link: link}
	err := d.templates.ExecuteTemplate(outputBuffer, name, env)
	if err != nil {
		return nil, newError("failed to execute template").Base(err)
	}
	return outputBuffer.Bytes(), nil
}

func (d *DefMatcher) ExecuteAll(link AbstractNonNativeLink) ([]byte, error) {
	outputBuffer := bytes.NewBuffer(nil)
	for _, loadedTemplates := range d.templates.Templates() {
		env := &ExecutionEnvironment{link: link}
		err := loadedTemplates.Execute(outputBuffer, env)
		if err != nil {
			outputBuffer.Reset()
		} else {
			break
		}
	}
	if outputBuffer.Len() == 0 {
		return nil, newError("failed to find a working template")
	}
	return outputBuffer.Bytes(), nil
}
