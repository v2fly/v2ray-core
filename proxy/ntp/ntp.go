//go:build !confonly
// +build !confonly

package ntp

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"context"
	"encoding/binary"
	"time"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/ntp/ntptime"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol/ntp"
	"github.com/v2fly/v2ray-core/v4/common/signal"
	"github.com/v2fly/v2ray-core/v4/common/task"
	"github.com/v2fly/v2ray-core/v4/features/policy"
	"github.com/v2fly/v2ray-core/v4/transport"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		h := new(Handler)
		if err := core.RequireFeatures(ctx, func(policyManager policy.Manager) error {
			return h.Init(config.(*Config), policyManager)
		}); err != nil {
			return nil, err
		}
		return h, nil
	}))
}

type Handler struct {
	timeout time.Duration
}

func (h *Handler) Init(config *Config, policyManager policy.Manager) error {
	h.timeout = policyManager.ForLevel(config.UserLevel).Timeouts.ConnectionIdle
	return nil
}

// Process implements proxy.Outbound.
func (h *Handler) Process(ctx context.Context, link *transport.Link, d internet.Dialer) error {
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, h.timeout)

	conn := net.NewConnection(net.ConnectionInputMulti(link.Writer))
	process := func() error {
		mb, err := link.Reader.ReadMultiBuffer()
		if err != nil {
			return err
		}
		now := ntp.ToNtpTime(ntptime.Now())
		defer buf.ReleaseMulti(mb)
		for _, buffer := range mb {
			request := new(ntp.Message)
			err = binary.Read(buffer, binary.BigEndian, request)
			buffer.Release()

			if err != nil {
				return err
			}

			request.Stratum = 3
			request.Precision = -25
			request.RootDispersion = 33
			request.SetMode(ntp.Server)
			request.SetLeap(ntp.LeapNoWarning)
			request.ReferenceTime = now
			request.OriginTime = request.TransmitTime
			request.ReceiveTime = now
			request.TransmitTime = ntp.ToNtpTime(ntptime.Now())

			err = binary.Write(conn, binary.BigEndian, request)

			if err != nil {
				return err
			}
		}

		timer.Update()
		return nil
	}

	if err := task.Run(ctx, func() (err error) {
		for err == nil {
			err = process()
		}
		return
	}); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}
