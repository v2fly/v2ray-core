package merge

import (
	"bytes"
	"encoding/json"
	"io"

	"v2ray.com/core/infra/conf/serial"
)

// FilesToJSON merges multiple jsons files into one json, accepts remote url, or local file path
func FilesToJSON(args []string) ([]byte, error) {
	m, err := FilesToMap(args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// BytesToJSON merges multiple json contents into one json.
func BytesToJSON(args [][]byte) ([]byte, error) {
	m, err := BytesToMap(args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// FilesToMap merges multiple json files into one map, accepts remote url, or local file path
func FilesToMap(args []string) (m map[string]interface{}, err error) {
	m, err = loadFiles(args)
	if err != nil {
		return nil, err
	}
	err = applyMergeRules(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// BytesToMap merges multiple json contents into one map.
func BytesToMap(args [][]byte) (m map[string]interface{}, err error) {
	m, err = loadBytes(args)
	if err != nil {
		return nil, err
	}
	err = applyMergeRules(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func applyMergeRules(m map[string]interface{}) error {
	sortSlicesInMap(m)
	err := mergeSameTag(m)
	if err != nil {
		return err
	}
	removeHelperFields(m)
	return nil
}

func loadFiles(args []string) (map[string]interface{}, error) {
	conf := make(map[string]interface{})
	for _, arg := range args {
		r, err := loadArg(arg)
		if err != nil {
			return nil, err
		}
		m, err := decode(r)
		if err != nil {
			return nil, err
		}
		if err = mergeMaps(conf, m); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func loadBytes(args [][]byte) (map[string]interface{}, error) {
	conf := make(map[string]interface{})
	for _, arg := range args {
		r := bytes.NewReader(arg)
		m, err := decode(r)
		if err != nil {
			return nil, err
		}
		if err = mergeMaps(conf, m); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func decode(r io.Reader) (map[string]interface{}, error) {
	c := make(map[string]interface{})
	err := serial.DecodeJSON(r, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
