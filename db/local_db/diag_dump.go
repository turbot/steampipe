package local_db

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"golang.org/x/sync/semaphore"
)

type DiagnosticDumpData struct {
	ProcessDump []string
	PortDump    []int
	Installed   []versionfile.InstalledVersion
	CallStack   string
	State       RunningDBInstanceInfo
	Version     string
}

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func (ps *PortScanner) Start(fromPort, toPort int, timeout time.Duration) []int {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	ports := []int{}
	for port := fromPort; port <= toPort; port++ {
		wg.Add(1)
		ps.lock.Acquire(context.TODO(), 1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			if scanPort(ps.ip, port, timeout) {
				ports = append(ports, port)
			}
		}(port)
	}
	return ports
}
func scanPort(ip string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return scanPort(ip, port, timeout)
		} else {
			return false
		}
	}
	conn.Close()
	return true
}
