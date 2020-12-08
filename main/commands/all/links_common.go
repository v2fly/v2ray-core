package all

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/link"
)

func writeFile(outdir, filename string, data []byte, filesMap map[string]string) error {
	if file, ok := filesMap[filename]; ok {
		// file exist
		rel, err := filepath.Rel(outdir, file)
		if err != nil {
			return err
		}
		hasher := md5.New()
		s, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		hasher.Write(s)
		fileMD5 := hex.EncodeToString(hasher.Sum(nil))
		hasher.Reset()
		hasher.Write(data)
		dataMD5 := hex.EncodeToString(hasher.Sum(nil))
		if fileMD5 != dataMD5 {
			fmt.Println("Updated:", rel)
			err = ioutil.WriteFile(file, data, 0644)
			if err != nil {
				return err
			}
		}
		delete(filesMap, filename)
		return nil
	}
	// file not exist
	file := filepath.Join(outdir, filename)
	fmt.Println("Added:", filename)
	return ioutil.WriteFile(file, data, 0644)
}

func getFilesMap(dir string) (map[string]string, error) {
	files, err := readDirRecursively(dir, []string{".json"})
	if err != nil {
		return nil, err
	}
	filesMap := make(map[string]string)
	for _, f := range files {
		filesMap[filepath.Base(f)] = f
	}
	return filesMap, nil
}

func asFileName(prefix, ps string) string {
	prefix = strings.TrimSpace(prefix)
	ps = strings.TrimSpace(ps)
	tag := ps
	if prefix != "" {
		tag = prefix + " - " + ps
	}
	reg := regexp.MustCompile(`([\\/:*?"<>|]|\s)+`)
	r := reg.ReplaceAll([]byte(tag), []byte(" "))
	return strings.TrimSpace(string(r))
}

// outbound2JSON converts vmess link to json string
func outbound2JSON(out *conf.OutboundDetourConfig, socketMark int32) ([]byte, error) {
	if socketMark != 0 {
		if out.StreamSetting == nil {
			out.StreamSetting = &conf.StreamConfig{}
		}
		out.StreamSetting.SocketSettings = &conf.SocketConfig{
			Mark: socketMark,
		}
	}
	type outConfig struct {
		OutboundConfigs []conf.OutboundDetourConfig `json:"outbounds"`
	}
	return json.Marshal(outConfig{
		OutboundConfigs: []conf.OutboundDetourConfig{
			*out,
		},
	})
}

func linkToJSON(link link.Link, tag string, socketMark int32) ([]byte, error) {
	out := link.ToOutbound()
	out.Tag = tag
	content, err := outbound2JSON(out, socketMark)
	if err != nil {
		return nil, err
	}
	return content, nil
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
