package main

import (
	"context"
	"encoding/binary"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// UDP 기반의 Ping Probe 프로그램으로, 패킷을 주기적으로 전송하고,
// 수신 응답을 처리하여 패킷 손실 및 지연 시간을 계산하는 애플리케이션입니다.

// listenAddr와 listenPort: 서버가 바인딩할 주소와 포트를 설정
// probeSizeBytes: 프로브 구조체의 크기
// retryTimeout: UDP 쓰기 작업의 타임아웃
// probeInterval: 패킷 전송 간격
var (
	listenAddr     = "0.0.0.0"
	listenPort     = 32767
	probeSizeBytes = 9
	maxReadBuffer  = 425984
	retryTimeout   = time.Second * 5
	probeInterval  = time.Second
)

func setupSigHandlers(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s", sig)
		cancel()
	}()
}

type probe struct {
	SeqNum uint8 // 각 프로브 패킷에 고유 번호를 부여하여 순서를 추적
	SendTS int64 // 프로브가 전송된 시점을 기록
}

// UDP 소켓에서 프로브 패킷을 읽고 손실 및 지연 시간(E2E latency)을 계산
// nextSeq: 다음에 기대하는 시퀀스 번호
// lost: 손실된 패킷 수를 추적
func receive(udpConn net.UDPConn) {
	log.Printf("Starting UDP Ping Receive")

	var nextSeq uint8
	var lost int
	for {

		p := &probe{}

		if err := binary.Read(&udpConn, binary.BigEndian, p); err != nil {
			return
		}

		log.Printf("Received probe %d", p.SeqNum)
		if p.SeqNum < nextSeq {
			log.Printf("Out of order packet seq/expected: %d/%d", p.SeqNum, nextSeq)
			lost -= 1
		} else if p.SeqNum > nextSeq {
			log.Printf("Out of order packet/expected: %d/%d", p.SeqNum, nextSeq)
			lost += int(p.SeqNum - nextSeq)
			nextSeq = p.SeqNum
		}

		latency := time.Now().UnixMilli() - p.SendTS
		log.Printf("E2E latency: %d ms", latency)
		log.Printf("Lost packets: %d", lost)
		nextSeq++
	}
}

func main() {
	server := flag.String("server", "127.0.0.1", "UDP Server IP")
	port := flag.Int("port", 32767, "UDP Server Port")
	flag.Parse()

	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(*server),
		Port: *port,
	}

	// 원격 서버와의 UDP 연결  설정
	// network: 문자열로, 사용할 네트워크 프로토콜을 나타냅니다. 일반적으로 "udp"또는 "udp4", "udp6"을 사용합니다.
	// lAddr: 로컬 주소를 나타내는 *net.UDPAddr타입입니다.
	// UDP 소켓이 데이터를 보낼 때 사용하는 로컬 주소를 지정합니다.
	// nil로 설정하면 운영 체제가 적절한 로컬 주소와 포트를 자동으로 할당합니다.
	// rAddr: 원격 주소를 나타내는 *net.UDPAddr타입입니다.
	// 데이터를 전송할 대상의 IP 주소와 포트를 지정합니다.
	// 반드시 설정해야 합니다.
	udpConn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer udpConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	setupSigHandlers(cancel)

	ticker := time.NewTicker(probeInterval)

	log.Printf("Starting UDP Ping Probe")
	go receive(*udpConn)

	var seq uint8
	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down UDP Ping Probe")
			return
		case <-ticker.C:
			log.Printf("Sending %d packets", probeSizeBytes)
			p := &probe{
				SeqNum: seq,
				SendTS: time.Now().UnixMilli(),
			}

			if err := udpConn.SetWriteDeadline(time.Now().Add(retryTimeout)); err != nil {
				log.Printf("Error setting write deadline: %v", err)
			}

			if err := binary.Write(udpConn, binary.BigEndian, p); err != nil {
				log.Printf("Error writing packet: %v", err)
			}

			seq++
		}
	}
}
