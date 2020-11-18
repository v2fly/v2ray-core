package peer

import (
	"sync"
)

type Latency interface {
	Value() uint64
}

type HasLatency interface {
	ConnectionLatency() Latency
	HandshakeLatency() Latency
}

type AverageLatency struct {
	access sync.Mutex
	value  uint64
}

func (al *AverageLatency) Update(newValue uint64) {
	if newValue == al.value {
		return
	}

	al.access.Lock()
	al.value = (al.value + newValue*2) / 3
	al.access.Unlock()
}

func (al *AverageLatency) Value() uint64 {
	return al.value
}
