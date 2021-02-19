package core

import "path/filepath"

//go:generate go get -v google.golang.org/protobuf/cmd/protoc-gen-go
//go:generate go get -v google.golang.org/grpc/cmd/protoc-gen-go-grpc
//go:generate go get -v github.com/gogo/protobuf/protoc-gen-gofast
//go:generate go run ./infra/vprotogen/

// ProtoFilesUsingProtocGenGoFast is the map of Proto files
// that use `protoc-gen-gofast` to generate pb.go files
var ProtoFilesUsingProtocGenGoFast = map[string]bool{filepath.Join("proxy", "vless", "encoding", "addons.proto"): true}
