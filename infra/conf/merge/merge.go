// Copyright 2020 Jebbs. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

/*
Package merge provides the capbility to merge multiple
JSON files or contents into one output.

Merge Rules:

- Simple values (string, number, boolean) are overwritten, others are merged
- Elements with same "tag" (or "_tag") in an array will be merged
- Add "_priority" property to array elements will help sort the

*/
package merge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
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
	err = applyRules(m)
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
	err = applyRules(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func loadFiles(args []string) (map[string]interface{}, error) {
	c := make(map[string]interface{})
	for _, arg := range args {
		r, err := cmdarg.LoadArg(arg)
		if err != nil {
			return nil, fmt.Errorf("fail to load %s: %s", arg, err)
		}
		m, err := decode(r)
		if err != nil {
			return nil, fmt.Errorf("fail to decode %s: %s", arg, err)
		}
		if err = mergeMaps(c, m); err != nil {
			return nil, fmt.Errorf("fail to merge %s: %s", arg, err)
		}
	}
	return c, nil
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
