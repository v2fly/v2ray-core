package stats_test

import (
	"context"
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/app/stats"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/features/stats"
)

func TestStatsCounter(t *testing.T) {
	raw, err := common.CreateObject(context.Background(), &Config{})
	common.Must(err)

	m := raw.(stats.Manager)
	c, err := m.RegisterCounter("test.counter")
	common.Must(err)

	if v := c.Add(1); v != 1 {
		t.Fatal("unpexcted Add(1) return: ", v, ", wanted ", 1)
	}

	if v := c.Set(0); v != 1 {
		t.Fatal("unexpected Set(0) return: ", v, ", wanted ", 1)
	}

	if v := c.Value(); v != 0 {
		t.Fatal("unexpected Value() return: ", v, ", wanted ", 0)
	}
}
