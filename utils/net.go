package utils

import (
	"fmt"
	"net"
	"time"
)

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

func IsPortBindable(port int) error {
	timeout := 5 * time.Millisecond
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)), timeout)
	if err != nil {
		return nil
	}
	if conn != nil {
		defer conn.Close()
		return fmt.Errorf("port %d is already in use", port)
	}
	return nil
}
