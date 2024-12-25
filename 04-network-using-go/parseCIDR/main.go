package main

import (
	"fmt"
	"net"
)

func main() {
	cidr, ipNet, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		panic(err)
	}

	fmt.Println(cidr, ipNet)

}
