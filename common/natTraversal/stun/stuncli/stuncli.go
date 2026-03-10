package stuncli

// Mostly machine generated code
import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	stunlib "github.com/v2fly/v2ray-core/v5/common/natTraversal/stun"
	vnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	"github.com/v2fly/v2ray-core/v5/proxy/socks"
)

var (
	server    *string
	server2   *string
	timeout   *int
	attempts  *int
	socks5udp *string
)

var cmdStunTest = &base.Command{
	UsageLine: "{{.Exec}} engineering stun-nat-type-discovery",
	Short:     "run STUN NAT type tests",
	Long: `
Run STUN NAT behavior discovery tests (RFC 5780) against a STUN server.

Tests NAT filtering, mapping, and hairpin behavior, then reports results.

The STUN server must support RFC 5780 (OTHER-ADDRESS and CHANGE-REQUEST)
for full test coverage.

Usage:
	{{.Exec}} engineering stun-test -server <host:port> [-server2 <host:port>] [-timeout <ms>] [-attempts <n>] [-socks5udp <host:port>]

Options:
	-server <host:port>
		The STUN server address (required)
	-server2 <host:port>
		A secondary STUN server address for cross-server mapping stability test
	-timeout <ms>
		Timeout per test in milliseconds (default: 3000)
	-attempts <n>
		Number of parallel requests per test for UDP loss resilience (default: 3)
	-socks5udp <host:port>
		SOCKS5 UDP relay address (skips TCP handshake, sends UDP directly)

Example:
	{{.Exec}} engineering stun-test -server stun.example.com:3478
	{{.Exec}} engineering stun-test -server stun.example.com:3478 -server2 stun2.example.com:3478
	{{.Exec}} engineering stun-test -server stun.example.com:3478 -socks5udp 127.0.0.1:1080
`,
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		server = fs.String("server", "", "STUN server address (host:port)")
		server2 = fs.String("server2", "", "secondary STUN server address (host:port)")
		timeout = fs.Int("timeout", 3000, "timeout per test in milliseconds")
		attempts = fs.Int("attempts", 3, "number of parallel requests per test")
		socks5udp = fs.String("socks5udp", "", "SOCKS5 UDP relay address (host:port)")
		return *fs
	}(),
	Run: executeStunTest,
}

func init() {
	engineering.AddCommand(cmdStunTest)
}

// socks5UDPConn wraps a PacketConn to encapsulate/decapsulate SOCKS5 UDP packets.
// All outgoing packets are wrapped in a SOCKS5 UDP header and sent to the relay.
// All incoming packets are unwrapped, with the real source address extracted from the header.
type socks5UDPConn struct {
	net.PacketConn
	relayAddr net.Addr
}

func (c *socks5UDPConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	udpAddr := addr.(*net.UDPAddr)
	dest := vnet.UDPDestination(vnet.IPAddress(udpAddr.IP), vnet.Port(udpAddr.Port))
	packet, err := socks.EncodeUDPPacketFromAddress(dest, p)
	if err != nil {
		return 0, err
	}
	defer packet.Release()
	_, err = c.PacketConn.WriteTo(packet.Bytes(), c.relayAddr)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *socks5UDPConn) ReadFrom(p []byte) (int, net.Addr, error) {
	// Allocate enough space for SOCKS5 header + payload
	rawBuf := make([]byte, len(p)+256)
	n, _, err := c.PacketConn.ReadFrom(rawBuf)
	if err != nil {
		return 0, nil, err
	}
	packet := buf.FromBytes(rawBuf[:n])
	req, err := socks.DecodeUDPPacket(packet)
	if err != nil {
		return 0, nil, err
	}
	// After DecodeUDPPacket, packet.Bytes() contains the payload
	dataN := copy(p, packet.Bytes())
	srcAddr := &net.UDPAddr{
		IP:   req.Address.IP(),
		Port: int(req.Port),
	}
	return dataN, srcAddr, nil
}

func natDependantTypeString(t stunlib.NATDependantType) string {
	switch t {
	case stunlib.Unknown:
		return "Unknown"
	case stunlib.Independent:
		return "Independent"
	case stunlib.EndpointDependent:
		return "Endpoint Dependent"
	case stunlib.EndpointPortDependent:
		return "Endpoint+Port Dependent"
	case stunlib.EndpointPortDependentPinned:
		return "Endpoint+Port Dependent (Pinned)"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func natYesOrNoString(t stunlib.NATYesOrNoUnknownType) string {
	switch t {
	case stunlib.NATYesOrNoUnknownType_Unknown:
		return "Unknown"
	case stunlib.NATYesOrNoUnknownType_Yes:
		return "Yes"
	case stunlib.NATYesOrNoUnknownType_No:
		return "No"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func executeStunTest(cmd *base.Command, args []string) {
	err := cmd.Flag.Parse(args)
	if err != nil {
		base.Fatalf("failed to parse flags: %v", err)
	}

	if *server == "" {
		base.Fatalf("-server is required")
	}

	host, portStr, err := net.SplitHostPort(*server)
	if err != nil {
		base.Fatalf("invalid server address %q: %v", *server, err)
	}

	ips, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		base.Fatalf("failed to resolve %q: %v", host, err)
	}

	port, err := net.LookupPort("udp", portStr)
	if err != nil {
		base.Fatalf("invalid port %q: %v", portStr, err)
	}

	serverAddr := &net.UDPAddr{IP: ips.IP, Port: port}

	// Resolve secondary STUN server address if provided
	var server2Addr *net.UDPAddr
	if *server2 != "" {
		host2, portStr2, err := net.SplitHostPort(*server2)
		if err != nil {
			base.Fatalf("invalid server2 address %q: %v", *server2, err)
		}
		ips2, err := net.ResolveIPAddr("ip", host2)
		if err != nil {
			base.Fatalf("failed to resolve server2 host %q: %v", host2, err)
		}
		port2, err := net.LookupPort("udp", portStr2)
		if err != nil {
			base.Fatalf("invalid server2 port %q: %v", portStr2, err)
		}
		server2Addr = &net.UDPAddr{IP: ips2.IP, Port: port2}
	}

	// Resolve SOCKS5 UDP relay address if provided
	var relayAddr *net.UDPAddr
	if *socks5udp != "" {
		rHost, rPortStr, err := net.SplitHostPort(*socks5udp)
		if err != nil {
			base.Fatalf("invalid socks5udp address %q: %v", *socks5udp, err)
		}
		rIPs, err := net.ResolveIPAddr("ip", rHost)
		if err != nil {
			base.Fatalf("failed to resolve socks5udp host %q: %v", rHost, err)
		}
		rPort, err := net.LookupPort("udp", rPortStr)
		if err != nil {
			base.Fatalf("invalid socks5udp port %q: %v", rPortStr, err)
		}
		relayAddr = &net.UDPAddr{IP: rIPs.IP, Port: rPort}
	}

	fmt.Printf("STUN server: %s\n", serverAddr)
	if server2Addr != nil {
		fmt.Printf("STUN server 2: %s\n", server2Addr)
	}
	if relayAddr != nil {
		fmt.Printf("SOCKS5 UDP relay: %s\n", relayAddr)
	}
	fmt.Printf("Timeout: %dms, Attempts: %d\n\n", *timeout, *attempts)

	newConn := func() (net.PacketConn, error) {
		conn, err := net.ListenPacket("udp", ":0")
		if err != nil {
			return nil, err
		}
		if relayAddr != nil {
			return &socks5UDPConn{PacketConn: conn, relayAddr: relayAddr}, nil
		}
		return conn, nil
	}

	var secondaryServer net.Addr
	if server2Addr != nil {
		secondaryServer = server2Addr
	}

	test := stunlib.NewNATTypeTest(
		newConn,
		serverAddr,
		secondaryServer,
		time.Duration(*timeout)*time.Millisecond,
		*attempts,
	)

	fmt.Println("Running tests...")
	if err := test.TestAll(); err != nil {
		base.Fatalf("test failed: %v", err)
	}

	fmt.Println()
	fmt.Println("=== NAT Behavior Test Results ===")
	fmt.Printf("  Filter Behaviour:  %s\n", natDependantTypeString(test.FilterBehaviour))
	fmt.Printf("  Mapping Behaviour: %s\n", natDependantTypeString(test.MappingBehaviour))
	fmt.Printf("  Hairpin Behaviour: %s\n", natYesOrNoString(test.HairpinBehaviour))
	fmt.Printf("  Stable Mapping on Secondary Server: %s\n", natYesOrNoString(test.StableMappingOnSecondaryServer))
	fmt.Println()
	fmt.Println("=== Derived Properties ===")
	fmt.Printf("  Preserve Source Port (Source NAT):     %s\n", natYesOrNoString(test.PreserveSourcePortWhenSourceNATMapping))
	fmt.Printf("  Single Source IP (Source NAT):         %s\n", natYesOrNoString(test.SingleSourceIPSourceNATMapping))
	fmt.Printf("  Preserve Source Addr (Dest NAT Reply): %s\n", natYesOrNoString(test.PreserveSourceIPPortWhenDestNATMapping))
	fmt.Println()
	fmt.Println("=== Source IPs ===")
	for _, ip := range test.SourceIPs {
		fmt.Printf("  %s\n", ip)
	}
}
