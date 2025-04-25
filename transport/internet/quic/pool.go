package quic

import (
	"sync"

	"github.com/ghxhy/v2ray-core/v5/common/bytespool"
)

var pool *sync.Pool

func init() {
	pool = bytespool.GetPool(2048)
}

func getBuffer() []byte {
	return pool.Get().([]byte)
}

func putBuffer(p []byte) {
	pool.Put(p) // nolint: staticcheck
}
