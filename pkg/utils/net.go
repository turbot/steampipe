package utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func GetFirstListenAddress(listenAddresses []string) string {
	listenAddress := strings.TrimSpace(listenAddresses[0])
	if listenAddress == "*" {
		listenAddress = "127.0.0.1"
	}
	return listenAddress
}

func ListenAddressesContainsOneOfAddresses(listenAddresses []string, addresses []string) bool {
	for i := range listenAddresses {
		listenAddress := strings.TrimSpace(listenAddresses[i])
		for j := range addresses {
			if addresses[j] == listenAddress {
				return true
			}
		}
	}
	return false
}

func LocalPublicAddresses() ([]string, error) {
	addresses := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				isToInclude := v.IP.IsGlobalUnicast() && ((v.IP.To4() != nil) || (v.IP.To16() != nil))
				if isToInclude {
					addresses = append(addresses, v.IP.String())
				}
			}

		}
	}

	return addresses, nil
}

func LocalLoopbackAddresses() ([]string, error) {
	addresses := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				isToInclude := v.IP.IsLoopback() && ((v.IP.To4() != nil) || (v.IP.To16() != nil))
				if isToInclude {
					addresses = append(addresses, v.IP.String())
				}
			}

		}
	}

	return addresses, nil
}

func IsPortBindable(host string, port int) error {
	timeout := 5 * time.Millisecond
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), timeout)
	if err != nil {
		return nil
	}
	if conn != nil {
		defer conn.Close()
		return fmt.Errorf("port %d is already in use", port)
	}
	return nil
}
