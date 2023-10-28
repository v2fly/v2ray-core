package packetparse

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

var (
	errNotIPPacket  = newError("not an IP packet")
	errNotUDPPacket = newError("not a UDP packet")
)

var nullDestination = net.UnixDestination(net.DomainAddress("null"))

func TryParseAsUDPPacket(packet []byte) (src, dst net.Destination, data []byte, err error) {
	parsedPacket := gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.DecodeOptions{
		Lazy:                     true,
		NoCopy:                   false,
		SkipDecodeRecovery:       false,
		DecodeStreamsAsDatagrams: false,
	})

	var srcIP net.Address
	var dstIP net.Address
	ipv4Layer := parsedPacket.Layer(layers.LayerTypeIPv4)

	if ipv4Layer == nil {
		parsedPacketAsIPv6 := gopacket.NewPacket(packet, layers.LayerTypeIPv6, gopacket.DecodeOptions{
			Lazy:                     true,
			NoCopy:                   false,
			SkipDecodeRecovery:       false,
			DecodeStreamsAsDatagrams: false,
		})
		ipv6Layer := parsedPacketAsIPv6.Layer(layers.LayerTypeIPv6)
		if ipv6Layer == nil {
			return nullDestination, nullDestination, nil, errNotIPPacket
		}
		ipv6 := ipv6Layer.(*layers.IPv6)
		srcIP = net.IPAddress(ipv6.SrcIP)
		dstIP = net.IPAddress(ipv6.DstIP)

		parsedPacket = parsedPacketAsIPv6
	} else {
		ipv4 := ipv4Layer.(*layers.IPv4)
		srcIP = net.IPAddress(ipv4.SrcIP)
		dstIP = net.IPAddress(ipv4.DstIP)
	}

	udpLayer := parsedPacket.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return nullDestination, nullDestination, nil, errNotUDPPacket
	}
	udp := udpLayer.(*layers.UDP)
	srcPort := net.Port(udp.SrcPort)
	dstPort := net.Port(udp.DstPort)

	src = net.UDPDestination(srcIP, srcPort)
	dst = net.UDPDestination(dstIP, dstPort)
	data = udp.Payload
	return // nolint: nakedret
}

func TryConstructUDPPacket(src, dst net.Destination, data []byte) ([]byte, error) {
	if src.Address.Family().IsIPv4() && dst.Address.Family().IsIPv4() {
		return constructIPv4UDPPacket(src, dst, data)
	}
	if src.Address.Family().IsIPv6() && dst.Address.Family().IsIPv6() {
		return constructIPv6UDPPacket(src, dst, data)
	}
	return nil, newError("not supported")
}

func constructIPv4UDPPacket(src, dst net.Destination, data []byte) ([]byte, error) {
	ipv4 := &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    src.Address.IP(),
		DstIP:    dst.Address.IP(),
		TTL:      64, // set TTL to a reasonable non-zero value to allow non-local routing
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(src.Port),
		DstPort: layers.UDPPort(dst.Port),
	}
	err := udp.SetNetworkLayerForChecksum(ipv4)
	if err != nil {
		return nil, err
	}
	buffer := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}, ipv4, udp, gopacket.Payload(data)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func constructIPv6UDPPacket(src, dst net.Destination, data []byte) ([]byte, error) {
	ipv6 := &layers.IPv6{
		Version:    6,
		NextHeader: layers.IPProtocolUDP,
		SrcIP:      src.Address.IP(),
		DstIP:      dst.Address.IP(),
		HopLimit:   64,
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(src.Port),
		DstPort: layers.UDPPort(dst.Port),
	}
	err := udp.SetNetworkLayerForChecksum(ipv6)
	if err != nil {
		return nil, err
	}
	buffer := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}, ipv6, udp, gopacket.Payload(data)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
