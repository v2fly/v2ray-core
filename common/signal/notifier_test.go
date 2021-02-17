package signal_test

import (
	"testing"

	. "github.com/v2fly/v2ray-core/v4/common/signal"
)

func TestNotifierSignal(t *testing.T) {
	n := NewNotifier()

	w := n.Wait()
	n.Signal()

	select {
	case <-w:
	default:
		t.Fail()
	}
}
