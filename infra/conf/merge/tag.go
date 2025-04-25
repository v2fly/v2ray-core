// Copyright 2020 Jebbs. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package merge

import (
	"strings"
)

func getTag(v map[string]interface{}, tagKeyOnly bool) string {
	if !tagKeyOnly {
		if field, ok := v["tag"]; ok {
			if t, ok := field.(string); ok {
				return t
			}
		}
	}
	if field, ok := v[tagKey]; ok {
		if t, ok := field.(string); ok {
			return t
		}
	}
	return ""
}

func mergeSameTag(s []interface{}, path string) ([]interface{}, error) {
	// from: [a,"",b,"",a,"",b,""]
	// to: [a,"",b,"",merged,"",merged,""]
	merged := &struct{}{}
	tagKeyOnly := false
	// See https://github.com/ghxhy/v2ray-core/issues/2981
	if strings.HasPrefix(path, ".dns.servers") {
		tagKeyOnly = true
	}
	for i, item1 := range s {
		map1, ok := item1.(map[string]interface{})
		if !ok {
			continue
		}
		tag1 := getTag(map1, tagKeyOnly)
		if tag1 == "" {
			continue
		}
		for j := i + 1; j < len(s); j++ {
			map2, ok := s[j].(map[string]interface{})
			if !ok {
				continue
			}
			tag2 := getTag(map2, tagKeyOnly)
			if tag1 == tag2 {
				s[j] = merged
				err := mergeMaps(map1, map2)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	// remove merged
	ns := make([]interface{}, 0)
	for _, item := range s {
		if item == merged {
			continue
		}
		ns = append(ns, item)
	}
	return ns, nil
}
