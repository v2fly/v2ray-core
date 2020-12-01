package external

//go:generate go run v2ray.com/core/common/errors/errorgen

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"v2ray.com/core/common/platform/ctlcmd"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/main/confloader"
)

// ConfigLoader is the protobuf config loader
func ConfigLoader(arg string) (out io.Reader, err error) {
	var data []byte
	switch {
	case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
		data, err = conf.FetchHTTPContent(arg)

	case arg == "stdin:":
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

// ExtConfigLoader calls v2ctl to load config
func ExtConfigLoader(args []string, reader io.Reader) (io.Reader, error) {
	buf, err := ctlcmd.Run(append([]string{"convert"}, args...), reader)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(buf.String()), nil
}

func init() {
	confloader.EffectiveConfigFileLoader = ConfigLoader
	confloader.EffectiveExtConfigLoader = ExtConfigLoader
}
