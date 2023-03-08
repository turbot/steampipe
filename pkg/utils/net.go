package utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func GetFirstListenAddress(listenAddresses string) string {
	return strings.TrimSpace(strings.Split(listenAddresses, ",")[0])
}

func LocalAddresses() ([]string, error) {
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
				isToInclude := v.IP.IsGlobalUnicast() && (v.IP.To4() != nil)
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
