// +build windows

package lile

import (
	"github.com/natefinch/npipe"
	"net"
)

func formatPlatformTestSeverAddress(uniquePortion string)(string) {
	return `\\.\pipe\` + uniquePortion
}

func getTestServerListener(address string)(net.Listener, error) {
	return npipe.Listen(address)
}

func dialTestServer(address string)(net.Conn, error) {
	return  npipe.Dial(address)
}
