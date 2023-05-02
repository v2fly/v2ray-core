package subscribe

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

var (
	client *http.Client // http 请求客户端
)

func init() {
	client = http.DefaultClient
	client.Timeout = 200 * time.Minute
	client.Transport = &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 500,
		MaxConnsPerHost:     500,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 100 * time.Second, // 指定tls握手超时
	}
}

// Get 函数用于发起一次 get 请求
func Get(url string) (resp string, err error) {
	log.Printf("visiting: %s\n", url)
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return "", errors.WithMessage(err, "http.NewRequest failed")
	}

	response, err := client.Do(req)
	if err != nil {
		return "", errors.WithMessage(err, "client.Do failed")
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("请求接口错误 statuscode=%v", response.StatusCode)
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.WithMessage(err, "ioutil.ReadAll failed")
	}

	return string(respBody), nil
}
