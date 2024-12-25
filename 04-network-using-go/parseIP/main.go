package main

import (
	"fmt"
	"net/netip"
)

/*
go에서 만든 IP 주소 타입
빅인디안 - 일반적으로 리틀엔디안을 사용하지만 네트워크에서는 빅 엔디안사용
*/
func main() {
	IPv4, err := netip.ParseAddr("224.0.0.1")
	if err != nil {
		panic(err)
	}

	// IPv4 address is Multicast를 출력한다.
	if IPv4.IsMulticast() {
		fmt.Println("IPv4 is multicast")
	}

	IPv6, err := netip.ParseAddr("FE80:F00D::1")
	if err != nil {
		panic(err)
	}

	// IPv6 address is Link Local Unicast를 출력한다.
	if IPv6.IsLinkLocalUnicast() {
		fmt.Println("IPv6 is link-local unicast")
	}
}
