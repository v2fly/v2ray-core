package ctlcmd

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/platform"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func Run(args []string, input io.Reader) (buf.MultiBuffer, error) {
	v2ctl := platform.GetToolLocation("v2ctl")
	if _, err := os.Stat(v2ctl); err != nil {
		return nil, newError("v2ctl doesn't exist").Base(err)
	}

	var errBuffer buf.MultiBufferContainer
	var outBuffer buf.MultiBufferContainer

	cmd := exec.Command(v2ctl, args...)
	cmd.Stderr = &errBuffer
	cmd.Stdout = &outBuffer
	cmd.SysProcAttr = getSysProcAttr()
	if input != nil {
		cmd.Stdin = input
	}

	if err := cmd.Start(); err != nil {
		return nil, newError("failed to start v2ctl").Base(err)
	}

	if err := cmd.Wait(); err != nil {
		msg := "failed to execute v2ctl"
		if errBuffer.Len() > 0 {
			msg += ": \n" + strings.TrimSpace(errBuffer.MultiBuffer.String())
		}
		return nil, newError(msg).Base(err)
	}

	// log stderr, info message
	if !errBuffer.IsEmpty() {
		newError("<v2ctl message> \n", strings.TrimSpace(errBuffer.MultiBuffer.String())).AtInfo().WriteToLog()
	}

	return outBuffer.MultiBuffer, nil
}
