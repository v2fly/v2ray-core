package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/pelletier/go-toml"
)

// FromTOML convert toml to json
func FromTOML(v []byte) ([]byte, error) {
	tomlReader := bytes.NewReader(v)
	jsonStr, err := jsonFromTomlReader(tomlReader)
	if err != nil {
		return nil, err
	}
	return []byte(jsonStr), nil
}

func jsonFromTomlReader(r io.Reader) (string, error) {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return "", err
	}
	return mapToJSON(tree)
}

func mapToJSON(tree *toml.Tree) (string, error) {
	treeMap := tree.ToMap()
	bytes, err := json.MarshalIndent(treeMap, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
