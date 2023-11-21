package base64urlline

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/common"
)

func newBase64URLLineParser() containers.SubscriptionContainerDocumentParser {
	return &parser{}
}

type parser struct{}

func (p parser) ParseSubscriptionContainerDocument(rawConfig []byte) (*containers.Container, error) {
	result := &containers.Container{}
	result.Kind = "Base64URLLine"
	result.Metadata = make(map[string]string)

	bodyDecoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(rawConfig))
	decoded, err := io.ReadAll(bodyDecoder)
	if err != nil {
		return nil, newError("failed to decode base64url body base64").Base(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(decoded))

	const maxCapacity int = 1024 * 256
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		result.ServerSpecs = append(result.ServerSpecs, containers.UnparsedServerConf{
			KindHint: "URL",
			Content:  scanner.Bytes(),
		})
	}
	return result, nil
}

func init() {
	common.Must(containers.RegisterParser("Base64URLLine", newBase64URLLineParser()))
}
