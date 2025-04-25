package log_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	. "github.com/ghxhy/v2ray-core/v5/common/log"
)

func TestFileLogger(t *testing.T) {
	f, err := os.CreateTemp("", "vtest")
	common.Must(err)
	path := f.Name()
	common.Must(f.Close())

	creator, err := CreateFileLogWriter(path)
	common.Must(err)

	handler := NewLogger(creator)
	handler.Handle(&GeneralMessage{Content: "Test Log"})
	time.Sleep(2 * time.Second)

	common.Must(common.Close(handler))

	f, err = os.Open(path)
	common.Must(err)
	defer f.Close()

	b, err := buf.ReadAllToBytes(f)
	common.Must(err)
	if !strings.Contains(string(b), "Test Log") {
		t.Fatal("Expect log text contains 'Test Log', but actually: ", string(b))
	}
}
