package core

//go:generate go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//go:generate go install -v github.com/gogo/protobuf/protoc-gen-gofast@latest
//go:generate go run ./infra/vprotogen/
//go:generate go run ./infra/errorgen/
