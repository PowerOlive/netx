// Package netx provides additional libraries that extend some of the behaviors
// in the net standard package.
package netx

import (
	"context"
	"net"
	"sync/atomic"
	"time"
)

var (
	dial           atomic.Value
	dialUDP        atomic.Value
	resolveTCPAddr atomic.Value
	resolveUDPAddr atomic.Value

	defaultDialTimeout = 1 * time.Minute
)

func init() {
	Reset()
}

// Dial is like DialTimeout using a default timeout of 1 minute.
func Dial(network string, addr string) (net.Conn, error) {
	return DialTimeout(network, addr, defaultDialTimeout)
}

func DialUDP(network string, laddr, raddr *net.UDPAddr) (net.Conn, error) {
	return dialUDP.Load().(func(string, *net.UDPAddr, *net.UDPAddr) (net.Conn, error))(network, laddr, raddr)
}

// DialTimeout dials the given addr on the given net type using the configured
// dial function, timing out after the given timeout.
func DialTimeout(network string, addr string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	conn, err := DialContext(ctx, network, addr)
	cancel()
	return conn, err
}

// DialContext dials the given addr on the given net type using the configured
// dial function, with the given context.
func DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	return dial.Load().(func(context.Context, string, string) (net.Conn, error))(ctx, network, addr)
}

// OverrideDial overrides the global dial function.
func OverrideDial(dialFN func(ctx context.Context, net string, addr string) (net.Conn, error)) {
	dial.Store(dialFN)
}

func OverrideUDPDial(dialFN func(net string, laddr, raddr *net.UDPAddr) (net.Conn, error)) {
	dialUDP.Store(dialFN)
}

// Resolve resolves the given tcp address using the configured resolve function.
func Resolve(network string, addr string) (*net.TCPAddr, error) {
	return resolveTCPAddr.Load().(func(string, string) (*net.TCPAddr, error))(network, addr)
}

// OverrideResolve overrides the global resolve function.
func OverrideResolve(resolveFN func(net string, addr string) (*net.TCPAddr, error)) {
	resolveTCPAddr.Store(resolveFN)
}

func ResolveUDP(network string, addr string) (*net.UDPAddr, error) {
	return resolveUDPAddr.Load().(func(string, string) (*net.UDPAddr, error))(network, addr)
}

func OverrideUDPResolve(resolveFN func(net string, addr string) (*net.UDPAddr, error)) {
	resolveUDPAddr.Store(resolveFN)
}

func SetDefaultUDPDial() {
	dialFN := func(net string, laddr, raddr *net.UDPAddr) (net.Conn, error) {
		conn, err := net.DialUDP(net, laddr, raddr)
		return conn, err
	}
	dialUDP.Store(dialFN)
}

// Reset resets netx to its default settings
func Reset() {
	var d net.Dialer
	OverrideDial(d.DialContext)
	SetDefaultUDPDial()
	OverrideResolve(net.ResolveTCPAddr)
	OverrideUDPResolve(net.ResolveUDPAddr)
}
