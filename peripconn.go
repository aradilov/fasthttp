package fasthttp

import (
	"fmt"
	"net"
	"sync"
)

type ipv6t [16]byte

type perIPConnCounter struct {
	pool sync.Pool
	lock sync.Mutex
	m    map[ipv6t]int
}

func (cc *perIPConnCounter) Register(ip net.IP) int {
	ipv6 := ip.To16()
	if nil == ipv6 {
		return 0
	}

	cc.lock.Lock()
	if cc.m == nil {
		cc.m = make(map[ipv6t]int)
	}
	n := cc.m[ipv6t(ipv6)] + 1
	cc.m[ipv6t(ipv6)] = n
	cc.lock.Unlock()
	return n
}

func (cc *perIPConnCounter) Unregister(ip net.IP) {

	ipv6 := ip.To16()
	if nil == ipv6 {
		return
	}

	cc.lock.Lock()
	if cc.m == nil {
		cc.lock.Unlock()
		panic("BUG: perIPConnCounter.Register() wasn't called")
	}
	n := cc.m[ipv6t(ipv6)] - 1
	if n < 0 {
		cc.lock.Unlock()
		panic(fmt.Sprintf("BUG: negative per-ip counter=%d for ip=%s", n, ip))
	}
	cc.m[ipv6t(ipv6)] = n
	cc.lock.Unlock()
}

type perIPConn struct {
	net.Conn

	ip               net.IP
	perIPConnCounter *perIPConnCounter
}

func acquirePerIPConn(conn net.Conn, ip net.IP, counter *perIPConnCounter) *perIPConn {
	v := counter.pool.Get()
	if v == nil {
		v = &perIPConn{
			perIPConnCounter: counter,
		}
	}
	c := v.(*perIPConn)
	c.Conn = conn
	c.ip = ip
	return c
}

func releasePerIPConn(c *perIPConn) {
	c.Conn = nil
	c.perIPConnCounter.pool.Put(c)
}

func (c *perIPConn) Close() error {
	err := c.Conn.Close()
	c.perIPConnCounter.Unregister(c.ip)
	releasePerIPConn(c)
	return err
}

func getConnIP6(c net.Conn) net.IP {
	addr := c.RemoteAddr()
	ipAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return net.IPv4zero
	}
	return ipAddr.IP.To16()
}

func ip2uint32(ip net.IP) uint32 {
	if len(ip) != 4 {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func uint322ip(ip uint32) net.IP {
	b := make([]byte, 4)
	b[0] = byte(ip >> 24)
	b[1] = byte(ip >> 16)
	b[2] = byte(ip >> 8)
	b[3] = byte(ip)
	return b
}

func copyAddr(src net.Addr) net.Addr {

	if tcpaddr, ok := src.(*net.TCPAddr); ok {
		dst := *tcpaddr
		return &dst
	}

	if tcpaddr, ok := src.(*net.UDPAddr); ok {
		dst := *tcpaddr
		return &dst
	}

	return src

}
