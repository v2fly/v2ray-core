package vmess3p

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"v2ray.com/core/common"
	"v2ray.com/core/infra/link"
)

func init() {
	common.Must(link.RegisterParser(&link.Parser{
		Name:   "V2RayNG",
		Scheme: []string{"vmess"},
		Parse: func(input string) (link.Link, error) {
			return NewVnVmess(input)
		},
	}))
}

// ToNgLink converts to V2RayNG link string
func (v TPLink) ToNgLink() string {
	b, _ := json.Marshal(v)
	return "vmess://" + base64.StdEncoding.EncodeToString(b)
}

// NewVnVmess parses V2RayN vemss link to VmessLink
func NewVnVmess(vmess string) (*TPLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}

	b64 := vmess[8:]
	b, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}

	v := &TPLink{}
	if err := json.Unmarshal(b, v); err != nil {
		return nil, err
	}
	v.OrigLink = vmess

	return v, nil
}
