package fasthttp

import (
	"net"
	"testing"
)

func TestIPxUint32(t *testing.T) {
	t.Parallel()

	testIPxUint32(t, 0)
	testIPxUint32(t, 10)
	testIPxUint32(t, 0x12892392)
}

func testIPxUint32(t *testing.T, n uint32) {
	ip := uint322ip(n)
	nn := ip2uint32(ip)
	if n != nn {
		t.Fatalf("Unexpected value=%d for ip=%s. Expected %d", nn, ip, n)
	}
}

func TestPerIPConnCounter(t *testing.T) {
	t.Parallel()

	var cc perIPConnCounter

	ip := net.ParseIP("127.0.0.1").To16()

	expectPanic(t, func() { cc.Unregister(ip) })

	for i := 1; i < 100; i++ {
		if n := cc.Register(ip); n != i {
			t.Fatalf("Unexpected counter value=%d. Expected %d", n, i)
		}
	}

	n := cc.Register(net.IPv4zero)
	if n != 1 {
		t.Fatalf("Unexpected counter value=%d. Expected 1", n)
	}

	for i := 1; i < 100; i++ {
		cc.Unregister(ip)
	}
	cc.Unregister(net.IPv4zero)

	expectPanic(t, func() { cc.Unregister(ip) })
	expectPanic(t, func() { cc.Unregister(net.IPv4zero) })

	n = cc.Register(ip)
	if n != 1 {
		t.Fatalf("Unexpected counter value=%d. Expected 1", n)
	}
	cc.Unregister(ip)
}

func expectPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expecting panic")
		}
	}()
	f()
}
