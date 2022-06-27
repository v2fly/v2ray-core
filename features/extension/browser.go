package extension

import (
	"io"
	"net/http"
)

type BrowserForwarder interface {
	DialWebsocket(url string, header http.Header) (io.ReadWriteCloser, error)
}

func BrowserForwarderType() interface{} {
	return (*BrowserForwarder)(nil)
}
