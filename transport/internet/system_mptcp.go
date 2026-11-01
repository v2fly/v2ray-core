package internet

import "github.com/v2fly/v2ray-core/v5/internal/transportdefaults"

func applyMPTCPConfig(socket interface{ SetMultipathTCP(bool) }, config *SocketConfig) {
	state := config.GetMptcp()
	if state == MPTCPState_AsIs && transportdefaults.MPTCPDisabled() {
		state = MPTCPState_Disable
	}

	switch state {
	case MPTCPState_Enable:
		socket.SetMultipathTCP(true)
	case MPTCPState_Disable:
		socket.SetMultipathTCP(false)
	}
}
