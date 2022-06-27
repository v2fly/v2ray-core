package net

import "net"

const (
	IPv4len = net.IPv4len
	IPv6len = net.IPv6len
)

var (
	CIDRMask        = net.CIDRMask
	Dial            = net.Dial
	DialTCP         = net.DialTCP
	DialUDP         = net.DialUDP
	DialUnix        = net.DialUnix
	FileConn        = net.FileConn
	Listen          = net.Listen
	ListenTCP       = net.ListenTCP
	ListenUDP       = net.ListenUDP
	ListenUnix      = net.ListenUnix
	LookupIP        = net.LookupIP
	ParseIP         = net.ParseIP
	ResolveUDPAddr  = net.ResolveUDPAddr
	ResolveUnixAddr = net.ResolveUnixAddr
	SplitHostPort   = net.SplitHostPort
)

type (
	Addr         = net.Addr
	AddrError    = net.AddrError
	Conn         = net.Conn
	Dialer       = net.Dialer
	Error        = net.Error
	IP           = net.IP
	IPMask       = net.IPMask
	IPNet        = net.IPNet
	ListenConfig = net.ListenConfig
	Listener     = net.Listener
	PacketConn   = net.PacketConn
	Resolver     = net.Resolver
	TCPAddr      = net.TCPAddr
	TCPConn      = net.TCPConn
	TCPListener  = net.TCPListener
	UDPAddr      = net.UDPAddr
	UDPConn      = net.UDPConn
	UnixAddr     = net.UnixAddr
	UnixConn     = net.UnixConn
	UnixListener = net.UnixListener
)
