package speed

import (
	"golang.org/x/time/rate"
	"sync"
	"v2ray.com/core/common/protocol"
)

// BucketHub is a speed limiter manager
type BucketHub struct {
	Users map[string]*rate.Limiter
	sync.RWMutex
}

// get NewBucketHub Instance
func NewBucketHub() *BucketHub {
	return newBucketHub
}

// The bucket that each user belongs to
func (b *BucketHub) GetUserBucket(u *protocol.MemoryUser, speed uint64) *rate.Limiter {
	if len(u.Email) > 0 && b.Users[u.Email] != nil {
		return b.Users[u.Email]
	}

	// 4 byte use one ticket, bursts 1M
	bucket := rate.NewLimiter(rate.Limit(speed / 4), 1024 * 1000)
	b.Lock()
	defer b.Unlock()
	b.Users[u.Email] = bucket
	return bucket
}

var newBucketHub *BucketHub

func init() {
	newBucketHub = new(BucketHub)
	newBucketHub.Users = make(map[string]*rate.Limiter)
}