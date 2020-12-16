package link

import (
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
)

func init() {
	common.Must(RegisterParser(&Parser{
		Name:   "Official",
		Scheme: []string{"vmess"},
		Parse:  ParseVmess,
	}))
}

// ParseVmess parses official vemss link to Link
func ParseVmess(vmess string) (Link, error) {
	// TODO: Official vmess:// parse support
	return nil, errors.New("not implemented")
}
