package internet

import (
	"net"
	"testing"

	"github.com/v2fly/v2ray-core/v5/internal/transportdefaults"
)

func TestMPTCPDefaultsAndOverrides(t *testing.T) {
	original := transportdefaults.MPTCPDisabled()
	t.Cleanup(func() {
		if original {
			transportdefaults.SetVersion("5.55.0")
		} else {
			transportdefaults.SetVersion("5.62.0")
		}
	})

	for _, godebug := range []string{"multipathtcp=0", "multipathtcp=1"} {
		t.Run(godebug, func(t *testing.T) {
			t.Setenv("GODEBUG", godebug)
			for _, version := range []string{"5.55.0", "5.61.9", "5.62.0", "6.0.0"} {
				t.Run(version, func(t *testing.T) {
					transportdefaults.SetVersion(version)
					var defaultListener net.ListenConfig
					var defaultDialer net.Dialer
					wantListener, wantDialer := defaultListener.MultipathTCP(), defaultDialer.MultipathTCP()
					if version == "5.55.0" || version == "5.61.9" {
						wantListener, wantDialer = false, false
					}

					for _, tc := range []struct {
						name         string
						config       *SocketConfig
						wantListener bool
						wantDialer   bool
					}{
						{"omitted-sockopt", nil, wantListener, wantDialer},
						{"omitted-mptcp", &SocketConfig{}, wantListener, wantDialer},
						{"enabled", &SocketConfig{Mptcp: MPTCPState_Enable}, true, true},
						{"disabled", &SocketConfig{Mptcp: MPTCPState_Disable}, false, false},
					} {
						t.Run(tc.name, func(t *testing.T) {
							var listener net.ListenConfig
							var dialer net.Dialer
							applyMPTCPConfig(&listener, tc.config)
							applyMPTCPConfig(&dialer, tc.config)
							if got := listener.MultipathTCP(); got != tc.wantListener {
								t.Errorf("listener MPTCP = %t, want %t", got, tc.wantListener)
							}
							if got := dialer.MultipathTCP(); got != tc.wantDialer {
								t.Errorf("dialer MPTCP = %t, want %t", got, tc.wantDialer)
							}
						})
					}
				})
			}
		})
	}
}
