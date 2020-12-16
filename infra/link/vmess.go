package link

import (
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
)

func init() {
	common.Must(RegisterParser(&Parser{
		Name:   "Official",
		Scheme: []string{"vmess"},
		Parse: func(input string) (Link, error) {
			return NewVmess(input)
		},
	}))
}

// NewVmess parses V2RayN vemss link to VmessLink
func NewVmess(vmess string) (Link, error) {
	// TODO: Official vmess:// parse support
	return nil, errors.New("not implemented")
}
