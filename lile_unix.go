// +build !windows

package lile

import (
	"net"
)

func formatPlatformTestSeverAddress(uniquePortion string)(string) {
	return "/tmp/" + uniquePortion
}

func getTestServerListener(address string)(net.Listener, error) {
	return net.Listen("unix", address)
}

func dialTestServer(address string)(net.Conn, error) {
	return  net.Dial("unix", address)
}
