package ntp_test

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol/ntp"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func TestQuery(t *testing.T) {
	server := net.Destination{
		Network: net.Network_UDP,
		Address: net.DomainAddress("time.cloudflare.com"),
		Port:    123,
	}
	conn, err := internet.DialSystem(context.Background(), server, nil)
	common.Must(err)

	message, ntpTime, err := ntp.Query(conn)
	common.Must(err)
	response := ntp.ParseTime(message, ntpTime)
	common.Must(response.Validate())

	println("offset:", response.ClockOffset.String())
}
