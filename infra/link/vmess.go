package link

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// VMessLink represents a parsed vmess link
type VMessLink struct {
	Ver      string      `json:"v"`
	Add      string      `json:"add"`
	Aid      interface{} `json:"aid"`
	Host     string      `json:"host"`
	ID       string      `json:"id"`
	Net      string      `json:"net"`
	Path     string      `json:"path"`
	Port     interface{} `json:"port"`
	Ps       string      `json:"ps"`
	TLS      string      `json:"tls"`
	Type     string      `json:"type"`
	OrigLink string      `json:"-"`
}

// IsEqual tests if this vmess link is equal to another
func (v *VMessLink) IsEqual(c *VMessLink) bool {
	realNet := func(n string) string {
		if n == "" {
			return "tcp"
		}
		return n
	}
	if realNet(v.Net) != realNet(c.Net) {
		return false
	}
	if fmt.Sprintf("%v", c.Port) != fmt.Sprintf("%v", v.Port) {
		return false
	}

	return v.Add == c.Add && v.Aid == c.Aid && v.Host == c.Host && v.ID == c.ID &&
		v.Path == c.Path && v.TLS == c.TLS && v.Type == c.Type
}

// LinkStr unmarshals VmessLink to string
func (v VMessLink) LinkStr(linkType string) string {
	switch strings.ToLower(linkType) {
	case "n", "ng", "nng":
		return v.asNgLink()
	case "rk", "rocket", "shadowrocket":
		return v.asRocketLink()
	case "quan", "quantumult":
		return v.asQuantumult()
	}

	return ""
}

// Tag return the tag of the link
func (v *VMessLink) Tag() string {
	return v.Ps
}

func (v VMessLink) String() string {
	return fmt.Sprintf("%s|%s|%v - (%s)", v.Net, v.Add, v.Port, v.Ps)
}

// Detail returns human readable string of VmessLink
func (v VMessLink) Detail() string {
	return fmt.Sprintf("Net: %s\nAddr: %s\nPort: %v\nUUID: %s\nType: %s\nTLS: %s\nPS: %s\n", v.Net, v.Add, v.Port, v.ID, v.Type, v.TLS, v.Ps)
}

func (v VMessLink) asNgLink() string {
	b, _ := json.Marshal(v)
	return "vmess://" + base64.StdEncoding.EncodeToString(b)
}

func (v VMessLink) asRocketLink() string {
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

func (v VMessLink) asQuantumult() string {
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
func NewQuanVmess(vmess string) (*VMessLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}
	b64 := vmess[8:]
	b, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}

	info := string(b)
	v := &VMessLink{}
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

// NewVnVmess parses V2RayN vemss link to VmessLink
func NewVnVmess(vmess string) (*VMessLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}

	b64 := vmess[8:]
	b, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}

	v := &VMessLink{}
	if err := json.Unmarshal(b, v); err != nil {
		return nil, err
	}
	v.OrigLink = vmess

	return v, nil
}

// NewRkVmess parses ShadowRocket vemss link string to VmessLink
func NewRkVmess(vmess string) (*VMessLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}
	url, err := url.Parse(vmess)
	if err != nil {
		return nil, err
	}
	link := &VMessLink{}
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

func base64Decode(b64 string) ([]byte, error) {
	b64 = strings.TrimSpace(b64)
	stdb64 := b64
	if pad := len(b64) % 4; pad != 0 {
		stdb64 += strings.Repeat("=", 4-pad)
	}

	b, err := base64.StdEncoding.DecodeString(stdb64)
	if err != nil {
		return base64.URLEncoding.DecodeString(b64)
	}
	return b, nil
}

// NewVmessLink parses vmess link to *VmessLink
// it will try different link formats
func NewVmessLink(vmess string) (*VMessLink, error) {
	var lk *VMessLink
	if o, nerr := NewVnVmess(vmess); nerr == nil {
		lk = o
	} else if o, rerr := NewRkVmess(vmess); rerr == nil {
		lk = o
	} else if o, qerr := NewQuanVmess(vmess); qerr == nil {
		lk = o
	} else {
		return nil, fmt.Errorf("%v, %v, %v", nerr, rerr, qerr)
	}
	return lk, nil
}
