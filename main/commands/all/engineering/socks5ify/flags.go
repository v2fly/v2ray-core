//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	"strings"
)

type bindFileFlags []bindFile

func (f *bindFileFlags) String() string {
	if f == nil {
		return ""
	}
	parts := make([]string, 0, len(*f))
	for _, item := range *f {
		parts = append(parts, item.Source+":"+item.Target)
	}
	return strings.Join(parts, ",")
}

func (f *bindFileFlags) Set(raw string) error {
	source, target, ok := strings.Cut(raw, ":")
	if !ok || source == "" || target == "" {
		return fmt.Errorf("bind-file must be source:target")
	}
	*f = append(*f, bindFile{Source: source, Target: target})
	return nil
}
