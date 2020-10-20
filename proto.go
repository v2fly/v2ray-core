package core

//go:generate go install -v google.golang.org/protobuf/cmd/protoc-gen-go
//go:generate go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc
//go:generate go install -v github.com/gogo/protobuf/protoc-gen-gofast
//go:generate go run ./infra/vprotogen/main.go

// ProtoFilesUsingProtocGenGoFast is the map of Proto files
// that use `protoc-gen-gofast` to generate pb.go files
var ProtoFilesUsingProtocGenGoFast = map[string]bool{"proxy/vless/encoding/addons.proto": true}
