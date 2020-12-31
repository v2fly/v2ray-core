package json

import (
	"encoding/json"

	"github.com/pelletier/go-toml"
)

// FromTOML convert toml to json
func FromTOML(v []byte) ([]byte, error) {
	m := make(map[string]interface{})
	if err := toml.Unmarshal(v, &m); err != nil {
		return nil, err
	}
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}
