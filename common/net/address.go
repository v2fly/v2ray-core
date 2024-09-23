package net

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"

	"github.com/golang/protobuf/jsonpb"
)

var (
	// LocalHostIP is a constant value for localhost IP in IPv4.
	LocalHostIP = IPAddress([]byte{127, 0, 0, 1})

	// AnyIP is a constant value for any IP in IPv4.
	AnyIP = IPAddress([]byte{0, 0, 0, 0})

	// LocalHostDomain is a constant value for localhost domain.
	LocalHostDomain = DomainAddress("localhost")

	// LocalHostIPv6 is a constant value for localhost IP in IPv6.
	LocalHostIPv6 = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

	// AnyIPv6 is a constant value for any IP in IPv6.
	AnyIPv6 = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
)

// AddressFamily is the type of address.
type AddressFamily byte

const (
	// AddressFamilyIPv4 represents address as IPv4
	AddressFamilyIPv4 = AddressFamily(0)

	// AddressFamilyIPv6 represents address as IPv6
	AddressFamilyIPv6 = AddressFamily(1)

	// AddressFamilyDomain represents address as Domain
	AddressFamilyDomain = AddressFamily(2)
)

// IsIPv4 returns true if current AddressFamily is IPv4.
func (af AddressFamily) IsIPv4() bool {
	return af == AddressFamilyIPv4
}

// IsIPv6 returns true if current AddressFamily is IPv6.
func (af AddressFamily) IsIPv6() bool {
	return af == AddressFamilyIPv6
}

// IsIP returns true if current AddressFamily is IPv6 or IPv4.
func (af AddressFamily) IsIP() bool {
	return af == AddressFamilyIPv4 || af == AddressFamilyIPv6
}

// IsDomain returns true if current AddressFamily is Domain.
func (af AddressFamily) IsDomain() bool {
	return af == AddressFamilyDomain
}

// Address represents a network address to be communicated with. It may be an IP address or domain
// address, not both. This interface doesn't resolve IP address for a given domain.
type Address interface {
	IP() net.IP     // IP of this Address
	Domain() string // Domain of this Address
	Family() AddressFamily
	String() string // String representation of this Address
}

func isAlphaNum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// ParseAddress parses a string into an Address. The return value will be an IPAddress when
// the string is in the form of IPv4 or IPv6 address, or a DomainAddress otherwise.
func ParseAddress(addr string) Address {
	// Handle IPv6 address in form as "[2001:4860:0:2001::68]"
	lenAddr := len(addr)
	if lenAddr > 0 && addr[0] == '[' && addr[lenAddr-1] == ']' {
		addr = addr[1 : lenAddr-1]
		lenAddr -= 2
	}

	if lenAddr > 0 && (!isAlphaNum(addr[0]) || !isAlphaNum(addr[len(addr)-1])) {
		addr = strings.TrimSpace(addr)
	}

	ip := net.ParseIP(addr)
	if ip != nil {
		return IPAddress(ip)
	}
	return DomainAddress(addr)
}

var bytes0 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// IPAddress creates an Address with given IP.
func IPAddress(ip []byte) Address {
	switch len(ip) {
	case net.IPv4len:
		var addr ipv4Address = [4]byte{ip[0], ip[1], ip[2], ip[3]}
		return addr
	case net.IPv6len:
		if bytes.Equal(ip[:10], bytes0) && ip[10] == 0xff && ip[11] == 0xff {
			return IPAddress(ip[12:16])
		}
		var addr ipv6Address = [16]byte{
			ip[0], ip[1], ip[2], ip[3],
			ip[4], ip[5], ip[6], ip[7],
			ip[8], ip[9], ip[10], ip[11],
			ip[12], ip[13], ip[14], ip[15],
		}
		return addr
	default:
		newError("invalid IP format: ", ip).AtError().WriteToLog()
		return nil
	}
}

// DomainAddress creates an Address with given domain.
func DomainAddress(domain string) Address {
	return domainAddress(domain)
}

type ipv4Address [4]byte

func (a ipv4Address) IP() net.IP {
	return net.IP(a[:])
}

func (ipv4Address) Domain() string {
	panic("Calling Domain() on an IPv4Address.")
}

func (ipv4Address) Family() AddressFamily {
	return AddressFamilyIPv4
}

func (a ipv4Address) String() string {
	return a.IP().String()
}

type ipv6Address [16]byte

func (a ipv6Address) IP() net.IP {
	return net.IP(a[:])
}

func (ipv6Address) Domain() string {
	panic("Calling Domain() on an IPv6Address.")
}

func (ipv6Address) Family() AddressFamily {
	return AddressFamilyIPv6
}

func (a ipv6Address) String() string {
	return "[" + a.IP().String() + "]"
}

type domainAddress string

func (domainAddress) IP() net.IP {
	panic("Calling IP() on a DomainAddress.")
}

func (a domainAddress) Domain() string {
	return string(a)
}

func (domainAddress) Family() AddressFamily {
	return AddressFamilyDomain
}

func (a domainAddress) String() string {
	return a.Domain()
}

// AsAddress translates IPOrDomain to Address.
func (d *IPOrDomain) AsAddress() Address {
	if d == nil {
		return nil
	}
	switch addr := d.Address.(type) {
	case *IPOrDomain_Ip:
		return IPAddress(addr.Ip)
	case *IPOrDomain_Domain:
		return DomainAddress(addr.Domain)
	}
	panic("Common|Net: Invalid address.")
}

// NewIPOrDomain translates Address to IPOrDomain
func NewIPOrDomain(addr Address) *IPOrDomain {
	switch addr.Family() {
	case AddressFamilyDomain:
		return &IPOrDomain{
			Address: &IPOrDomain_Domain{
				Domain: addr.Domain(),
			},
		}
	case AddressFamilyIPv4, AddressFamilyIPv6:
		return &IPOrDomain{
			Address: &IPOrDomain_Ip{
				Ip: addr.IP(),
			},
		}
	default:
		panic("Unknown Address type.")
	}
}

func (d *IPOrDomain) UnmarshalJSONPB(unmarshaler *jsonpb.Unmarshaler, bytes []byte) error {
	var ipOrDomain string
	if err := json.Unmarshal(bytes, &ipOrDomain); err != nil {
		return err
	}
	result := NewIPOrDomain(ParseAddress(ipOrDomain))
	d.Address = result.Address
	return nil
}

func (d *IPOrDomain) MarshalJSONPB(marshaler *jsonpb.Marshaler) ([]byte, error) {
	ipod := d.AsAddress().String()

	return json.Marshal(ipod)
}
