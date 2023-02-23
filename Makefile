PROTOS = $(patsubst ./%,%,$(shell find . -name "*.proto"))

deps: $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-go-grpc
	go mod download

protoc:
	@ protoc --go_out=. --go-grpc_out=. $(PROTOS)