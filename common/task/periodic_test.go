package task_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v4/common"
	. "github.com/v2fly/v2ray-core/v4/common/task"
)

func TestPeriodicTaskStop(t *testing.T) {
	var value uint64
	task := &Periodic{
		Interval: time.Second * 2,
		Execute: func() error {
			atomic.AddUint64(&value, 1)
			return nil
		},
	}
	common.Must(task.Start())
	time.Sleep(time.Second * 5)
	common.Must(task.Close())
	value1 := atomic.LoadUint64(&value)
	if value1 != 3 {
		t.Fatal("expected 3, but got ", value1)
	}

	time.Sleep(time.Second * 4)
	value2 := atomic.LoadUint64(&value)
	if value2 != 3 {
		t.Fatal("expected 3, but got ", value2)
	}

	common.Must(task.Start())
	time.Sleep(time.Second * 3)
	value3 := atomic.LoadUint64(&value)
	if value3 != 5 {
		t.Fatal("Expected 5, but ", value3)
	}
	common.Must(task.Close())
}
