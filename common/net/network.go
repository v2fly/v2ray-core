package net

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/jsonpb"
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

func (nl *NetworkList) UnmarshalJSONPB(unmarshaler *jsonpb.Unmarshaler, bytes []byte) error {
	var networkStrList []string
	if err := json.Unmarshal(bytes, &networkStrList); err == nil {
		nl.Network = ParseNetworkStringList(networkStrList)
		return nil
	}

	var networkStr string
	if err := json.Unmarshal(bytes, &networkStr); err == nil {
		strList := strings.Split(networkStr, ",")
		nl.Network = ParseNetworkStringList(strList)
		return nil
	}

	return newError("unknown format of a string list: " + string(bytes))
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

func ParseNetworkStringList(strList []string) []Network {
	list := make([]Network, len(strList))
	for idx, str := range strList {
		list[idx] = ParseNetwork(str)
	}

	return list
}
