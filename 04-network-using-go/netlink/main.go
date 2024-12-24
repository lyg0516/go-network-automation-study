package main

import (
	"log"
	"net"

	"github.com/jsimonetti/rtnetlink/rtnl"
)

/*
	이 코드는 루프백 인터페이스(lo)를 찾은 뒤, 해당 인터페이스를 Down 상태로 만들었다가
	다시 Up 상태로 복구하는 작업을 수행합니다.
*/

func main() {
	// rtnl 패키지를 사용해 Netlink 소켓에 연결을 엽니다.
	// nil을 전달하면 기본 네트워크 네임스페이스와 연결됩니다.
	conn, err := rtnl.Dial(nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 시스템의 모든 네트워크 인터페이스 정보를 가져옵니다.
	// 변환된 links는 각 네트워크 인터페이스를 나타내는 객체들의 리스트입니다.
	links, err := conn.Links()

	var loopback *net.Interface

	// 모든 네트워크를 순회하면서 이름이 "lo"인 인터페이스를 찾습니다.
	for _, l := range links {
		if l.Name == "lo" {
			loopback = l
			log.Printf("Name: %s, Flags: %s", l.Name, l.Flags)
		}
	}

	// 루프백 인터페이스를 Down 상태로 전환합니다.
	conn.LinkDown(loopback)
	// 루프백 인터페이스의 인덱스를 이용해 해당 인터페이스 정보를 다시 가져옵니다.
	loopback, _ = conn.LinkByIndex(loopback.Index)
	log.Printf("Name: %s, Flags: %s", loopback.Name, loopback.Flags)

	// 루프백 인터페이스를 다시 Up 상태로 전환합니다.
	conn.LinkUp(loopback)
	loopback, _ = conn.LinkByIndex(loopback.Index)
	log.Printf("Name: %s, Flags: %s", loopback.Name, loopback.Flags)
}
