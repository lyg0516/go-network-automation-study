package main

import (
	"fmt"
	"github.com/c-robinson/iplib"
	"net"
	"sort"
)

func main() {
	// 문자열 "192.0.2.1"을 net.IP타입으로 변환합니다.
	// 결과는 IP 주소 객체 192.0.2.1입니다.
	IP := net.ParseIP("192.0.2.1")

	// iplib.NextIP(IP)는 입력된 IP 주소의 다음 주소를 계산합니다.
	nextIP := iplib.NextIP(IP)

	// iplib.IncrementIPBY는 지정된 숫자만큼 IP 주소를 증가시킵니다.
	incrIP := iplib.IncrementIPBy(nextIP, 19)

	// iplib.DeltaIP는 두 IP 주소 사이의 차이를 계산합니다.
	fmt.Println(iplib.DeltaIP(IP, incrIP))

	// iplib.CompareIPS는 두 IP 주소를 비교합니다.
	fmt.Println(iplib.CompareIPs(IP, incrIP))

	iplist := []net.IP{incrIP, nextIP, IP}
	fmt.Println(iplist)

	sort.Sort(iplib.ByIP(iplist))
	fmt.Println(iplist)

	// iplib.NewNet4는 IP 네트워크 객체를 생성
	// "198.51.100.0/24" 네트워크를 생성
	n4 := iplib.NewNet4(net.ParseIP("198.51.100.0"), 24)
	fmt.Println("Total IP addresses: ", n4.Count())

	fmt.Println("First three IPs: ", n4.Enumerate(3, 0))
	fmt.Println("First IP: ", n4.FirstAddress())
	fmt.Println("Last IP: ", n4.LastAddress())
}
