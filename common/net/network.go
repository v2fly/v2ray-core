package net

import (
	"strings"
)

func (n Network) SystemString() string {
	switch n {
	case Network_TCP:
		return "tcp"
	case Network_UDP:
		return "udp"
	case Network_UNIX:
		return "unix"
	default:
		return "unknown"
	}
}

// HasNetwork returns true if the network list has a certain network.
func HasNetwork(list []Network, network Network) bool {
	for _, value := range list {
		if value == network {
			return true
		}
	}
	return false
}

func ParseNetwork(net string) Network {
	switch strings.ToLower(net) {
	case "tcp":
		return Network_TCP
	case "udp":
		return Network_UDP
	case "unix":
		return Network_UNIX
	default:
		return Network_Unknown
	}
}

func ParseNetworks(netlist string) []Network {
	strlist := strings.Split(netlist, ",")
	nl := make([]Network, len(strlist))
	for idx, network := range strlist {
		nl[idx] = ParseNetwork(network)
	}

	return nl
}
