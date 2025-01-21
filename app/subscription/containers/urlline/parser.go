package urlline

import (
	"bufio"
	"bytes"
	"net/url"
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/common"
)

func newURLLineParser() containers.SubscriptionContainerDocumentParser {
	return &parser{}
}

type parser struct{}

func (p parser) ParseSubscriptionContainerDocument(rawConfig []byte) (*containers.Container, error) {
	result := &containers.Container{}
	result.Kind = "URLLine"
	result.Metadata = make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewReader(rawConfig))

	const maxCapacity int = 1024 * 256
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	parsedLine := 0
	failedLine := 0

	for scanner.Scan() {
		content := scanner.Text()
		content = strings.TrimSpace(content)
		if strings.HasPrefix(content, "#") {
			continue
		}
		if strings.HasPrefix(content, "//") {
			continue
		}
		_, err := url.Parse(content)
		if err != nil {
			failedLine++
			continue
		} else {
			parsedLine++
		}
		result.ServerSpecs = append(result.ServerSpecs, containers.UnparsedServerConf{
			KindHint: "URL",
			Content:  scanner.Bytes(),
		})
	}

	if failedLine > parsedLine || parsedLine == 0 {
		return nil, newError("failed to parse as URLLine").Base(newError("too many failed lines"))
	}

	return result, nil
}

func init() {
	common.Must(containers.RegisterParser("URLLine", newURLLineParser()))
}
