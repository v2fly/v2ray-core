package vmess3p

import (
	"encoding/base64"
	"fmt"
	"strings"

	"v2ray.com/core/common"
	"v2ray.com/core/infra/link"
)

func init() {
	common.Must(link.RegisterParser(&link.Parser{
		Name:   "Quantumult",
		Scheme: []string{"vmess"},
		Parse: func(input string) (link.Link, error) {
			return NewQuanVmess(input)
		},
	}))
}

// ToQuantumult converts to Quantumult link string
func (v TPLink) ToQuantumult() string {
	/*
	   let obfs = `,obfs=${jsonConf.net === 'ws' ? 'ws' : 'http'},obfs-path="${jsonConf.path || '/'}",obfs-header="Host:${jsonConf.host || jsonConf.add}[Rr][Nn]User-Agent:${ua}"`
	   let quanVmess  = `${jsonConf.ps} = vmess,${jsonConf.add},${jsonConf.port},${method},"${jsonConf.id}",over-tls=${jsonConf.tls === 'tls' ? 'true' : 'false'},certificate=1${jsonConf.type === 'none' && jsonConf.net !== 'ws' ? '' : obfs},group=${group}`
	*/

	method := "aes-128-gcm"
	vbase := fmt.Sprintf("%s = vmess,%s,%s,%s,\"%s\",over-tls=%v,certificate=1", v.Ps, v.Add, v.Port, method, v.ID, v.TLS == "tls")

	var obfs string
	if (v.Net == "ws" || v.Net == "http") && (v.Type == "none" || v.Type == "") {
		if v.Path == "" {
			v.Path = "/"
		}
		if v.Host == "" {
			v.Host = v.Add
		}
		obfs = fmt.Sprintf(`,obfs=ws,obfs-path="%s",obfs-header="Host:%s[Rr][Nn]User-Agent:Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/16A5366a"`, v.Path, v.Host)
	}

	vbase += obfs
	vbase += ",group=Fndroid"
	return "vmess://" + base64.URLEncoding.EncodeToString([]byte(vbase))
}

// NewQuanVmess parses Quantumult vemss link to VmessLink
func NewQuanVmess(vmess string) (*TPLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}
	b64 := vmess[8:]
	b, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}

	info := string(b)
	v := &TPLink{}
	v.OrigLink = vmess
	v.Ver = "2"

	psn := strings.SplitN(info, " = ", 2)
	if len(psn) != 2 {
		return nil, fmt.Errorf("part error: %s", info)
	}
	v.Ps = psn[0]
	params := strings.Split(psn[1], ",")
	v.Add = params[1]
	v.Port = params[2]
	v.ID = strings.Trim(params[4], "\"")
	v.Aid = "0"
	v.Net = "tcp"
	v.Type = "none"

	if len(params) > 4 {
		for _, pkv := range params[5:] {
			kvp := strings.SplitN(pkv, "=", 2)
			if kvp[0] == "over-tls" && kvp[1] == "true" {
				v.TLS = "tls"
			}

			if kvp[0] == "obfs" && kvp[1] == "ws" {
				v.Net = "ws"
			}

			if kvp[0] == "obfs" && kvp[1] == "http" {
				v.Type = "http"
			}

			if kvp[0] == "obfs-path" {
				v.Path = strings.Trim(kvp[1], "\"")
			}

			if kvp[0] == "obfs-header" {
				hd := strings.Trim(kvp[1], "\"")
				for _, hl := range strings.Split(hd, "[Rr][Nn]") {
					if strings.HasPrefix(hl, "Host:") {
						host := hl[5:]
						if host != v.Add {
							v.Host = host
						}
						break
					}
				}
			}
		}
	}

	return v, nil
}
