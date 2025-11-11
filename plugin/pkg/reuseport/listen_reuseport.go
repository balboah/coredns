//go:build go1.11 && (aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd)

package reuseport

import (
	"context"
	"net"
	"syscall"

	"github.com/coredns/coredns/plugin/pkg/log"

	"golang.org/x/sys/unix"
)

func reusePortControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			log.Warningf("Failed to set SO_REUSEPORT on socket: %s", err)
		}
	})
}

// Listen announces on the local network address. See net.Listen for more information.
// If SO_REUSEPORT is available it will be set on the socket.
func Listen(network, addr string) (net.Listener, error) {
	return listenWithControl(network, addr, reusePortControl)
}

// ListenWithControl behaves like Listen but also applies the provided control
// function to the socket. The SO_REUSEPORT option is still configured first.
func ListenWithControl(network, addr string, extra func(network, address string, c syscall.RawConn) error) (net.Listener, error) {
	return listenWithControl(network, addr, chainControls(reusePortControl, extra))
}

// ListenPacket announces on the local network address. See net.ListenPacket for more information.
// If SO_REUSEPORT is available it will be set on the socket.
func ListenPacket(network, addr string) (net.PacketConn, error) {
	return listenPacketWithControl(network, addr, reusePortControl)
}

// ListenPacketWithControl behaves like ListenPacket but also applies the
// provided control function to the socket. The SO_REUSEPORT option is still
// configured first.
func ListenPacketWithControl(network, addr string, extra func(network, address string, c syscall.RawConn) error) (net.PacketConn, error) {
	return listenPacketWithControl(network, addr, chainControls(reusePortControl, extra))
}

func listenWithControl(network, addr string, control func(network, address string, c syscall.RawConn) error) (net.Listener, error) {
	lc := net.ListenConfig{}
	if control != nil {
		lc.Control = control
	}
	return lc.Listen(context.Background(), network, addr)
}

func listenPacketWithControl(network, addr string, control func(network, address string, c syscall.RawConn) error) (net.PacketConn, error) {
	lc := net.ListenConfig{}
	if control != nil {
		lc.Control = control
	}
	return lc.ListenPacket(context.Background(), network, addr)
}
