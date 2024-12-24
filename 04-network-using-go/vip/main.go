package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsimonetti/rtnetlink/rtnl"
	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/packet"
)

/*
	이 코드는 VIP를 특정 네트워크 이너페이스에 추가하고, 주기적으로
	GARP(Gratuitous ARP) 메시지를 전송하여 VIP를 알리는 프로그램입니다.
	시그널을 기다리는 고루틴과
	3초 마다 GARP를 보내는 고루틴이 있으며 시그널이 들어오면 컨텍스트를 cancel해 main을 종료시킵니다.
*/

const VIP1 = "198.51.100.1/32"

/*
IP, netlink 소켓, 네트워크 인터페이스, layer2 socket
*/
type vip struct {
	IP      string
	netlink *rtnl.Conn
	intf    *net.Interface
	l2Sock  *packet.Conn
}

func setupSigHandlers(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go func() {
		sig := <-sigs
		log.Printf("Received syscall: %+v", sig)
		cancel()
	}()

}

// VIP 추가 및 삭제
// netlink API를 활용해 네트워크 인터페이스에 VIP 할당및 제거
// rtnl.MustPareAddr: IP 주소를 문자열로 입력받아 Netlink가 처리할 수 있는 IP 주소 객체로 변환
func (c *vip) removeVIP() error {
	err := c.netlink.AddrDel(c.intf, rtnl.MustParseAddr(c.IP))
	if err != nil {
		return fmt.Errorf("could not del address: %s", err)
	}
	return nil
}

func (c *vip) addVIP() error {
	err := c.netlink.AddrAdd(c.intf, rtnl.MustParseAddr(c.IP))
	if err != nil {
		return fmt.Errorf("could not add address: %s", err)
	}
	return nil
}

// Ethernet 프레임을 네트워크 인터페이스로 전송하는 역할
func (c *vip) emitFrame(frame *ethernet.Frame) error {
	// Ethernet 프레임 객체를 바이너리 데이터로 직렬화
	b, err := frame.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error serializing frame: %s", err)
	}

	// 전송 대상의 MAC 주소를 의미하며, Broadcast 설정
	addr := &packet.Addr{HardwareAddr: ethernet.Broadcast}

	// L2 소켓을 사용하여 직렬화된 Ethernet 프레임을 지정된 MAC 주소로 전송합니다.
	// 네트워크 인터페이스에 직접 접근하여 데이터 링크 계층 수준에서 패킷을 전송합니다.
	if _, err := c.l2Sock.WriteTo(b, addr); err != nil {
		return fmt.Errorf("emitFrame failed: %s", err)
	}

	log.Println("GARP sent")
	return nil
}

// GARP(Gratuitous ARP)는 자신의 IP를 브로드캐스트로 알리는 ARP 패킷입니다.
func (c *vip) sendGARP() error {
	ip, _, err := net.ParseCIDR(c.IP)
	if err != nil {
		return fmt.Errorf("error parsing IP: %s", err)
	}

	// Gratuitous ARP 패킷을 생성합니다.
	// Gratuitous ARP는 특별한 목적의 ARP 요청 또는 응답으로, IP 주소를 네트워크에 브로드캐스트 하기 위해 사용됩니다.
	arpPayload, err := arp.NewPacket(
		arp.OperationReply,  // ARP 응답 패킷으로 설정
		c.intf.HardwareAddr, // 네트워크 인터페이스의 MAC주소. 이 MAC 주소가 패킷의 출발지와 목적지 모두에 사용됩니다.
		ip,                  // srcIP
		c.intf.HardwareAddr, //dstHW
		ip,                  //dstIP
	)
	if err != nil {
		return fmt.Errorf("error building ARP packet: %s", err)
	}

	// ARP 패킷 직렬화
	// MarshalBinary: ARP 패킷 객체를 이진 데이터로 직렬화하여 실제 네트워크에 전송 가능한 형태로 변환합니다.
	// ARP 패킷을 이진 데이터로 변환하여 Ethernet 프레임의 페이로드로 활용
	arpBinary, err := arpPayload.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error serializing ARP packet: %s", err)
	}

	// ARP 패킷을 Ethernet 프레임의 페이로드로 캡슐화
	ethFrame := &ethernet.Frame{
		Destination: ethernet.Broadcast,
		Source:      c.intf.HardwareAddr,
		EtherType:   ethernet.EtherTypeARP,
		Payload:     arpBinary,
	}

	return c.emitFrame(ethFrame)
}

func main() {

	// -intf 플래그를 통해 VIP를 추가할 네트워크 인터페이스를 지정합니다.
	// 지정된 네트워크 인터페이스가 없으면 프로그램은 종료됩니다.
	// 플래그는 go 프로그램을 실행할때 외부에서 변수를 받기 위한 패키지
	intfStr := flag.String("intf", "", "VIP interface")
	flag.Parse()

	if *intfStr == "" {
		log.Fatal("Please provide -intf flag")
	}

	// 네트워크 인터페이스 이름을 기반으로 해당 인터페이스를 가져옵니다.
	// VIP를 이 네트워크 인터페이스에 바인딩합니다.
	netIntf, err := net.InterfaceByName(*intfStr)
	if err != nil {
		log.Fatalf("interface not found: %s", err)
	}

	// Netlink 연결: rtnl 객체를 사용해 Linux Netlink API와 통신합니다. 이를 통해 네트워크 인터페이스에 IP를 추가/제거합니다.
	rtnl, err := rtnl.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer rtnl.Close()

	// Raw 이더넷 소켓: 네트워크 레벨의 Raw 소켓을 열어 GARP 메시지를 브로드캐스트합니다.
	// 패킷 소켓을 생성하고 네트워크 인터페이스를 통해 특정 유형의 네트워크 패킷을 수신하는 데 사용
	// proto Protocol: packet.Raw 모든 패킷을 수신
	// snaplen int: 최대 캡처할 패킷 크기를 지정 0이면 기본 값 사용
	// bpfFilter []byte: BPF 필터를 사용해 패킷을 필터링합니다.
	// 리턴: *Socket: 생성된 패킷 소켓입니다. 이후 이 소켓을 통해 패킷을 읽거나 처리할 수 있습니다.
	ethSocket, err := packet.Listen(netIntf, packet.Raw, 0, nil)
	if err != nil {
		log.Printf("failed to ListenPacket: %v", err)
	}
	defer ethSocket.Close()

	// ctx는 VIP 관리 작업의 전체 생명 주기를 제어합니다.
	// cancel() 호출 시 ctx.Done() 채널이 닫히고, 이를 통해 종료 신호를 전파합니다.
	ctx, cancel := context.WithCancel(context.Background())

	// 일반적으로 SIGINT(Ctrl + C)나 SIGTERM 같은 시스템 신호를 처리하도록 설정합니다.
	setupSigHandlers(cancel)

	// VIP 객체 초기화
	v := &vip{
		IP:      VIP1,
		intf:    netIntf,
		netlink: rtnl,
		l2Sock:  ethSocket,
	}

	// VIP를 지정된 네트워크 인터페이스에 추가합니다.
	err = v.addVIP()
	if err != nil {
		log.Fatalf("failed to add VIP: %s", err)
	}

	// 3초마다 GARP 패킷을 전송하기 위한 타이머를 생성합니다.
	timer := time.NewTicker(3 * time.Second)

	// 컨텍스트가 종료되면 VIP를 네트워크 인터페이스 제거하고 프로그램을 종료합니다.
	for {
		select {
		case <-ctx.Done():
			if err := v.removeVIP(); err != nil {
				log.Fatalf("failed to remove VIP: %s", err)
			}

			log.Printf("Cleanup complete")
			return
		case <-timer.C:
			if err := v.sendGARP(); err != nil {
				log.Printf("failed to send GARP: %s", err)
				cancel()
			}
		}
	}

}
