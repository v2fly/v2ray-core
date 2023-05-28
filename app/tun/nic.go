package tun

import (
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func CreateNIC(nicID tcpip.NICID, linkEndpoint stack.LinkEndpoint) StackOption {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(nicID, linkEndpoint,
			stack.NICOptions{
				Disabled: false,
				QDisc:    nil,
			}); err != nil {
			return newError("failed to create NIC:", err)
		}
		return nil
	}
}
