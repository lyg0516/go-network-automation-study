package main

import (
	"fmt"
	"github.com/yl2chen/cidranger"
	"net"
)

/*
cidr의 contains를 써도 되지만
트라이 구조의 cidranger를 사용하면 성능좋은 포함여부 검사가 가능하다.
*/
func main() {

	// Prefix Compression Trie 자료구조를 사용하여 CIDR 관리
	ranger := cidranger.NewPCTrieRanger()

	IPs := []string{
		"100.64.0.0/16",
		"127.0.0.0/8",
		"172.16.0.0/16",
		"192.0.2.0/24",
		"192.0.2.0/24",
		"192.0.2.0/25",
		"192.0.2.127/25",
	}

	for _, prefix := range IPs {
		_, ipv4Net, err := net.ParseCIDR(prefix)
		if err != nil {
			panic(err)
		}
		// CIDR 네트워크를 ranger에 추가
		ranger.Insert(cidranger.NewBasicRangerEntry(*ipv4Net))
	}

	// 특정 IP가 ranger에 등록된 네트워크에 포함되는지 검사
	checkIP := "127.0.0.1"
	ok, err := ranger.Contains(net.ParseIP(checkIP))

	fmt.Printf("Does the range contains: %s?: %v\n", checkIP, ok)

	netIP := "192.0.2.18"

	nets, err := ranger.ContainingNetworks(net.ParseIP(netIP))
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nNetworks that contain IP address %s => \n", netIP)
	for _, e := range nets {
		n := e.Network()
		fmt.Println("\t", n.String())
	}

}
