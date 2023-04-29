all: v2ray

v2ray:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o output/v2ray -ldflags "-s -w"

.PHONY: v2ray
