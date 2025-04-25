package core

//go:generate env GOBIN=$PWD go install -v github.com/golang/mock/mockgen@latest
//go:generate ./mockgen -package mocks -destination testing/mocks/io.go -mock_names Reader=Reader,Writer=Writer io Reader,Writer
//go:generate ./mockgen -package mocks -destination testing/mocks/log.go -mock_names Handler=LogHandler github.com/ghxhy/v2ray-core/v5/common/log Handler
//go:generate ./mockgen -package mocks -destination testing/mocks/mux.go -mock_names ClientWorkerFactory=MuxClientWorkerFactory github.com/ghxhy/v2ray-core/v5/common/mux ClientWorkerFactory
//go:generate ./mockgen -package mocks -destination testing/mocks/dns.go -mock_names Client=DNSClient github.com/ghxhy/v2ray-core/v5/features/dns Client
//go:generate ./mockgen -package mocks -destination testing/mocks/outbound.go -mock_names Manager=OutboundManager,HandlerSelector=OutboundHandlerSelector github.com/ghxhy/v2ray-core/v5/features/outbound Manager,HandlerSelector
//go:generate ./mockgen -package mocks -destination testing/mocks/proxy.go -mock_names Inbound=ProxyInbound,Outbound=ProxyOutbound github.com/ghxhy/v2ray-core/v5/proxy Inbound,Outbound
