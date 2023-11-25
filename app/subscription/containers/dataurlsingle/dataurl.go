package dataurlsingle

import (
	"bytes"
	"strings"

	"github.com/vincent-petithory/dataurl"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/common"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func newSingularDataURLParser() containers.SubscriptionContainerDocumentParser {
	return &parser{}
}

type parser struct{}

func (p parser) ParseSubscriptionContainerDocument(rawConfig []byte) (*containers.Container, error) {
	dataURL, err := dataurl.Decode(bytes.NewReader(rawConfig))
	if err != nil {
		return nil, newError("unable to decode dataURL").Base(err)
	}
	if dataURL.MediaType.Type != "application" {
		return nil, newError("unsupported media type: ", dataURL.MediaType.Type)
	}
	if !strings.HasPrefix(dataURL.MediaType.Subtype, "vnd.v2ray.subscription-singular") {
		return nil, newError("unsupported media subtype: ", dataURL.MediaType.Subtype)
	}
	result := &containers.Container{}
	result.Kind = "DataURLSingle"
	result.Metadata = make(map[string]string)
	result.ServerSpecs = append(result.ServerSpecs, containers.UnparsedServerConf{
		KindHint: "",
		Content:  dataURL.Data,
	})
	return result, nil
}

func init() {
	common.Must(containers.RegisterParser("DataURLSingle", newSingularDataURLParser()))
}
