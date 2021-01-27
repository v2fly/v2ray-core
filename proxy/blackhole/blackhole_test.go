package blackhole_test

import (
	"context"
	"testing"

	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/blackhole"
	"v2ray.com/core/transport"
	"v2ray.com/core/transport/pipe"
)

func TestBlackHoleHTTPResponse(t *testing.T) {
	handler, err := blackhole.New(context.Background(), &blackhole.Config{
		Response: serial.ToTypedMessage(&blackhole.HTTPResponse{}),
	})
	common.Must(err)

	reader, writer := pipe.New(pipe.WithoutSizeLimit())

	var readerError = make(chan error)
	var mb buf.MultiBuffer
	go func() {
		b, e := reader.ReadMultiBuffer()
		mb = b
		readerError <- e
	}()

	link := transport.Link{
		Reader: reader,
		Writer: writer,
	}
	common.Must(handler.Process(context.Background(), &link, nil))
	common.Must(<-readerError)
	if mb.IsEmpty() {
		t.Error("expect http response, but nothing")
	}
}
