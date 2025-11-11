//go:build !go1.11 || (!aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd)

package reuseport

import (
	"net"
	"syscall"
)

// Listen is a wrapper around net.Listen.
func Listen(network, addr string) (net.Listener, error) { return net.Listen(network, addr) }

// ListenWithControl ignores the provided control function on platforms where
// SO_REUSEPORT isn't available and falls back to a standard net.Listen call.
func ListenWithControl(network, addr string, _ func(network, address string, c syscall.RawConn) error) (net.Listener, error) {
	return Listen(network, addr)
}

// ListenPacket is a wrapper around net.ListenPacket.
func ListenPacket(network, addr string) (net.PacketConn, error) {
	return net.ListenPacket(network, addr)
}

// ListenPacketWithControl ignores the provided control function on platforms
// where SO_REUSEPORT isn't available.
func ListenPacketWithControl(network, addr string, _ func(network, address string, c syscall.RawConn) error) (net.PacketConn, error) {
	return ListenPacket(network, addr)
}
