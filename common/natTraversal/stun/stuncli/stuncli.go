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
	server                 *string
	server2                *string
	timeout                *int
	attempts               *int
	socks5udp              *string
	detectBuggyNATMapping  *bool
	mappingLifetimeMaxIdle *int
	mappingLifetimeStart   *int
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
	-detect-buggy-nat-mapping
		Send additional probes to detect NAT mapping behavior that depends on packet arrival ordering
	-mapping-lifetime-max-idle <ms>
		Optional maximum idle window for UDP mapping lifetime probing (default: 0, disabled)
	-mapping-lifetime-start-idle <ms>
		Initial idle window for UDP mapping lifetime probing (default: 1000)

Example:
	{{.Exec}} engineering stun-test -server stun.example.com:3478
	{{.Exec}} engineering stun-test -server stun.example.com:3478 -server2 stun2.example.com:3478
	{{.Exec}} engineering stun-test -server stun.example.com:3478 -socks5udp 127.0.0.1:1080
	{{.Exec}} engineering stun-test -server stun.example.com:3478 -mapping-lifetime-max-idle 60000
`,
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		server = fs.String("server", "", "STUN server address (host:port)")
		server2 = fs.String("server2", "", "secondary STUN server address (host:port)")
		timeout = fs.Int("timeout", 3000, "timeout per test in milliseconds")
		attempts = fs.Int("attempts", 3, "number of parallel requests per test")
		socks5udp = fs.String("socks5udp", "", "SOCKS5 UDP relay address (host:port)")
		detectBuggyNATMapping = fs.Bool("detect-buggy-nat-mapping", false, "send additional probes to detect NAT mapping behavior that depends on packet arrival ordering")
		mappingLifetimeMaxIdle = fs.Int("mapping-lifetime-max-idle", 0, "optional maximum idle window for UDP mapping lifetime probing in milliseconds")
		mappingLifetimeStart = fs.Int("mapping-lifetime-start-idle", 1000, "initial idle window for UDP mapping lifetime probing in milliseconds")
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
	case stunlib.EndpointPortDependentStaticPinned:
		return "Endpoint+Port Dependent (Static Pinned)"
	case stunlib.EndpointPortDependentMappingPinned:
		return "Endpoint+Port Dependent (Mapping Pinned)"
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

func mappingLifetimeString(lowerBound, upperBound time.Duration) string {
	switch {
	case upperBound > 0 && lowerBound > 0:
		return fmt.Sprintf("[%v, %v)", lowerBound, upperBound)
	case upperBound > 0:
		return fmt.Sprintf("< %v", upperBound)
	case lowerBound > 0:
		return fmt.Sprintf(">= %v", lowerBound)
	default:
		return "Unknown"
	}
}

func mappingLifetimeEstimateString(estimate stunlib.MappingLifetimeEstimate) string {
	if !estimate.Supported {
		return "Not supported"
	}
	return mappingLifetimeString(estimate.LowerBound, estimate.UpperBound)
}

func printMappingLifetimeResults(test *stunlib.NATTypeTest) {
	fmt.Println("=== Optional Mapping Lifetime Probe ===")
	fmt.Println("  Probe Model: fresh UDP socket per idle interval")
	fmt.Printf("  Primary Server:     %s\n", mappingLifetimeString(test.MappingLifetimeLowerBound, test.MappingLifetimeUpperBound))
	fmt.Printf("  Other Address+Port: %s\n", mappingLifetimeEstimateString(test.MappingLifetimeOtherAddr))
	fmt.Printf("  Secondary Server:   %s\n", mappingLifetimeEstimateString(test.MappingLifetimeSecondary))
	fmt.Printf("  Independent:        %s\n", mappingLifetimeEstimateString(test.MappingLifetimeIndependent))
	fmt.Println()
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
	fmt.Printf("Detect buggy NAT mapping: %t\n\n", *detectBuggyNATMapping)
	if *mappingLifetimeMaxIdle > 0 {
		fmt.Printf(
			"Optional mapping lifetime probe: max idle %dms, initial idle %dms\n\n",
			*mappingLifetimeMaxIdle,
			*mappingLifetimeStart,
		)
	}

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
	test.DetectBuggyNATMapping = *detectBuggyNATMapping

	fmt.Println("Running tests...")
	if err := test.TestAll(); err != nil {
		base.Fatalf("test failed: %v", err)
	}
	if *mappingLifetimeMaxIdle > 0 {
		fmt.Println("Running optional mapping lifetime probe...")
		if err := test.TestMappingLifetime(
			time.Duration(*mappingLifetimeMaxIdle)*time.Millisecond,
			time.Duration(*mappingLifetimeStart)*time.Millisecond,
		); err != nil {
			base.Fatalf("mapping lifetime probe failed: %v", err)
		}
	}

	fmt.Println()
	fmt.Println("=== NAT Behavior Test Results ===")
	fmt.Printf("  Filter Behaviour:  %s\n", natDependantTypeString(test.FilterBehaviour))
	fmt.Printf("  Mapping Behaviour: %s\n", natDependantTypeString(test.MappingBehaviour))
	fmt.Printf("  Hairpin Behaviour: %s\n", natYesOrNoString(test.HairpinBehaviour))
	fmt.Printf("  Stable Mapping on Secondary Server: %s\n", natYesOrNoString(test.StableMappingOnSecondaryServer))
	fmt.Printf("  Incorrect Response Origin: %s\n", natYesOrNoString(test.IncorrectResponseOrigin))
	fmt.Println()
	fmt.Println("=== Derived Properties ===")
	fmt.Printf("  Preserve Source Port (Source NAT):     %s\n", natYesOrNoString(test.PreserveSourcePortWhenSourceNATMapping))
	fmt.Printf("  Single Source IP (Source NAT):         %s\n", natYesOrNoString(test.SingleSourceIPSourceNATMapping))
	fmt.Printf("  Preserve Source Addr (Dest NAT Reply): %s\n", natYesOrNoString(test.PreserveSourceIPPortWhenDestNATMapping))
	fmt.Println()
	if *mappingLifetimeMaxIdle > 0 {
		printMappingLifetimeResults(test)
	}
	fmt.Println("=== Source IPs ===")
	for _, ip := range test.SourceIPs {
		fmt.Printf("  %s\n", ip)
	}
}
