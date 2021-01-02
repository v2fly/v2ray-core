// +build android

package net

import (
	"context"
	"fmt"
	"net"
)

func SetDefaultResolver(ns []byte) string {
	if ns == nil {
		return ""
	}
	defaultNS := fmt.Sprintf("%d.%d.%d.%d", ns[0], ns[1], ns[2], ns[3])
	var dialer net.Dialer
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context context.Context, _, _ string) (net.Conn, error) {
			conn, err := dialer.DialContext(context, "udp", defaultNS+":53")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	return defaultNS
}
