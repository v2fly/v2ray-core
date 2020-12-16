package vmess3p

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"v2ray.com/core/common"
	"v2ray.com/core/infra/link"
)

func init() {
	common.Must(link.RegisterParser(&link.Parser{
		Name:   "ShadowRocket",
		Scheme: []string{"vmess"},
		Parse: func(input string) (link.Link, error) {
			return NewRkVmess(input)
		},
	}))
}

// ToRocketLink converts to ShadowRocket link string
func (v TPLink) ToRocketLink() string {
	mhp := fmt.Sprintf("%s:%s@%s:%s", v.Type, v.ID, v.Add, v.Port)
	qs := url.Values{}
	qs.Add("remarks", v.Ps)
	if v.Net == "ws" {
		qs.Add("obfs", "websocket")
	}
	if v.Host != "" {
		qs.Add("obfsParam", v.Host)
	}
	if v.Path != "" {
		qs.Add("path", v.Host)
	}
	if v.TLS == "tls" {
		qs.Add("tls", "1")
	}

	url := url.URL{
		Scheme:   "vmess",
		Host:     base64.URLEncoding.EncodeToString([]byte(mhp)),
		RawQuery: qs.Encode(),
	}

	return url.String()
}

// NewRkVmess parses ShadowRocket vemss link string to VmessLink
func NewRkVmess(vmess string) (*TPLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}
	url, err := url.Parse(vmess)
	if err != nil {
		return nil, err
	}
	link := &TPLink{}
	link.Ver = "2"
	link.OrigLink = vmess

	b64 := url.Host
	b, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}

	mhp := strings.SplitN(string(b), ":", 3)
	if len(mhp) != 3 {
		return nil, fmt.Errorf("vmess unreconized: method:host:port -- %v", mhp)
	}
	// mhp[0] is the encryption method
	link.Port = mhp[2]
	idadd := strings.SplitN(mhp[1], "@", 2)
	if len(idadd) != 2 {
		return nil, fmt.Errorf("vmess unreconized: id@addr -- %v", idadd)
	}
	link.ID = idadd[0]
	link.Add = idadd[1]
	link.Aid = "0"

	vals := url.Query()
	if v := vals.Get("remarks"); v != "" {
		link.Ps = v
	}
	if v := vals.Get("path"); v != "" {
		link.Path = v
	}
	if v := vals.Get("tls"); v == "1" {
		link.TLS = "tls"
	}
	if v := vals.Get("obfs"); v != "" {
		switch v {
		case "websocket":
			link.Net = "ws"
		case "none":
			link.Net = "tcp"
			link.Type = "none"
		}
	}
	if v := vals.Get("obfsParam"); v != "" {
		link.Host = v
	}

	return link, nil
}
