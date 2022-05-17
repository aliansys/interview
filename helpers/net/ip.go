package net

import (
	"net"
)

func IpFromAddressString(address string) (net.IP, error) {
	ip, _, err := net.SplitHostPort(address)
	if err != nil {
		return net.IP{}, err
	}

	userIP := net.ParseIP(ip)
	return userIP, nil
}
