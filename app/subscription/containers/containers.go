package containers

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type UnparsedServerConf struct {
	KindHint string
	Content  []byte
}

type Container struct {
	Kind        string
	Metadata    map[string]string
	ServerSpecs []UnparsedServerConf
}

type SubscriptionContainerDocumentParser interface {
	ParseSubscriptionContainerDocument(rawConfig []byte) (*Container, error)
}

var knownParsers = make(map[string]SubscriptionContainerDocumentParser)

func RegisterParser(kind string, parser SubscriptionContainerDocumentParser) error {
	if _, found := knownParsers[kind]; found {
		return newError("parser already registered for kind ", kind)
	}
	knownParsers[kind] = parser
	return nil
}
