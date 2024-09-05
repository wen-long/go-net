// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv6_test

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"testing"

	"github.com/wen-long/go-net/internal/iana"
	"github.com/wen-long/go-net/ipv6"
	"github.com/wen-long/go-net/nettest"
)

func TestConnInitiatorPathMTU(t *testing.T) {
	switch runtime.GOOS {
	case "fuchsia", "hurd", "js", "nacl", "plan9", "windows", "wasip1", "zos":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	if _, err := nettest.RoutedInterface("ip6", net.FlagUp|net.FlagLoopback); err != nil {
		t.Skip("ipv6 is not enabled for loopback interface")
	}

	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan bool)
	go acceptor(t, ln, done)

	c, err := net.Dial("tcp6", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if pmtu, err := ipv6.NewConn(c).PathMTU(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("path mtu for %v: %v", c.RemoteAddr(), pmtu)
	}

	<-done
}

func TestConnResponderPathMTU(t *testing.T) {
	switch runtime.GOOS {
	case "fuchsia", "hurd", "js", "nacl", "plan9", "windows", "wasip1", "zos":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if _, err := nettest.RoutedInterface("ip6", net.FlagUp|net.FlagLoopback); err != nil {
		t.Skip("ipv6 is not enabled for loopback interface")
	}

	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan bool)
	go connector(t, "tcp6", ln.Addr().String(), done)

	c, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if pmtu, err := ipv6.NewConn(c).PathMTU(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("path mtu for %v: %v", c.RemoteAddr(), pmtu)
	}

	<-done
}

func TestPacketConnChecksum(t *testing.T) {
	if !nettest.SupportsIPv6() {
		t.Skip("ipv6 is not supported")
	}
	if !nettest.SupportsRawSocket() {
		t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	c, err := net.ListenPacket(fmt.Sprintf("ip6:%d", iana.ProtocolOSPFIGP), "::") // OSPF for IPv6
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	p := ipv6.NewPacketConn(c)
	offset := 12 // see RFC 5340

	for _, toggle := range []bool{false, true} {
		if err := p.SetChecksum(toggle, offset); errors.Is(err, ipv6.ErrNotImplemented) {
			t.Skipf("setting checksum not supported: %v", err)
		} else if err != nil {
			if toggle {
				t.Fatalf("ipv6.PacketConn.SetChecksum(%v, %v) failed: %v", toggle, offset, err)
			} else {
				// Some platforms never allow to disable the kernel
				// checksum processing.
				t.Logf("ipv6.PacketConn.SetChecksum(%v, %v) failed: %v", toggle, offset, err)
			}
		}
		if on, offset, err := p.Checksum(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("kernel checksum processing enabled=%v, offset=%v", on, offset)
		}
	}
}
