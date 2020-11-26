package merge

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"v2ray.com/core/common/buf"
)

// JSONs merges multiple jsons into one.
// It accepts local files, URLs
func JSONs(files []string) ([]byte, error) {
	m, err := JSONsToMap(files)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// JSONsToMap merges multiple jsons into one map.
// It accepts []string for URLs, files, [][]byte for json contents
func JSONsToMap(args interface{}) (map[string]interface{}, error) {
	conf := make(map[string]interface{})
	switch v := args.(type) {
	case []string:
		for _, arg := range v {
			r, err := loadArg(arg)
			if err != nil {
				return nil, err
			}
			m, err := jsonToMap(r)
			if err != nil {
				return nil, err
			}
			if err = mergeMaps(conf, m); err != nil {
				return nil, err
			}
		}
	case [][]byte:
		for _, arg := range v {
			r := bytes.NewReader(arg)
			m, err := jsonToMap(r)
			if err != nil {
				return nil, err
			}
			if err = mergeMaps(conf, m); err != nil {
				return nil, err
			}
		}
	default:
		return nil, errors.New("unsupport args for JSONsToMap")
	}
	sortSlicesInMap(conf)
	removePriorityKey(conf)
	return conf, nil
}

// loadArg loads one arg, maybe an remote url, or local file path
func loadArg(arg string) (out io.Reader, err error) {
	var data []byte
	switch {
	case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
		data, err = fetchHTTPContent(arg)
	case (arg == "stdin:"):
		data, err = ioutil.ReadAll(os.Stdin)
	default:
		data, err = ioutil.ReadFile(arg)
	}
	if err != nil {
		return
	}
	out = bytes.NewBuffer(data)
	return
}

// fetchHTTPContent dials https for remote content
func fetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, newError("invalid URL: ", target).Base(err)
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, newError("invalid scheme: ", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, newError("failed to dial to ", target).Base(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newError("unexpected HTTP status code: ", resp.StatusCode)
	}

	content, err := buf.ReadAllToBytes(resp.Body)
	if err != nil {
		return nil, newError("failed to read HTTP response").Base(err)
	}

	return content, nil
}
