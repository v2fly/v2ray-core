package subscribe

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	Convey("测试函数 Get", t, func() {
		resp, err := Get("https://1.amazom.group/link/ylTi7Cf1y6rG2K0U?sub=3")
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeEmpty)
		var vmessBytes []byte
		vmessBytes, err = base64.StdEncoding.DecodeString(resp)
		So(err, ShouldBeNil)
		vmesses := strings.Split(string(vmessBytes), "\n")
		for _, vmess := range vmesses {
			if !strings.HasPrefix(vmess, "vmess://") {
				continue
			}
			vmess = strings.TrimPrefix(vmess, "vmess://")
			vmessBytes, err = base64.StdEncoding.DecodeString(vmess)
			So(err, ShouldBeNil)
			vmessConfig := new(VmessConfig)
			err = json.Unmarshal(vmessBytes, &vmessConfig)
			So(err, ShouldBeNil)
		}
	})
}
