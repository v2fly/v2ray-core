package vmess3p_test

import (
	"testing"

	"v2ray.com/core/infra/link"
	"v2ray.com/core/infra/link/vmess3p"
)

func TestParse(t *testing.T) {
	lk := &vmess3p.TPLink{
		Ps: "test",
	}
	l, err := link.Parse(lk.ToNgLink())
	if err != nil {
		t.Error(err)
	}
	if lk2, _ := l.(*vmess3p.TPLink); !lk2.IsEqual(lk) {
		t.Errorf("expected: %v, actual: %v", lk, lk2)
	}
	l, err = link.Parse(lk.ToRocketLink())
	if err != nil {
		t.Error(err)
	}
	l, err = link.Parse(lk.ToQuantumult())
	if err != nil {
		t.Error(err)
	}
}
