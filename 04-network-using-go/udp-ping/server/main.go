package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 이 코드는 UDP Ping 서버로 동작합니다. 클라이언트에서 보낸 UDP 패킷을 그대로 다시 클라이언트에게
// 반환(Echo) 하는 프로그램입니다.

// 전역 변수
// 이 변수들은 서버의 네트워크 동작을 조정하는 데 사용됩니다.
var (
	listenAddr     = "0.0.0.0"
	listenPort     = 32767
	probeSizeBytes = 9
	maxReadBuffer  = 425984
	retryTimeout   = time.Second * 5
	probeInterval  = time.Second
)

// 시그널 핸들러 설정
// 이 함수는 운영체네 시그널을 처리합니다.
// 시그널이 발생하면 cancel()을 호출하여 서버를 종료합니다.
func setupSigHandlers(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go func() {
		sig := <-sigs
		log.Println("Got signal:", sig)
		cancel()
	}()
}

type probe struct {
	SeqNum uint8
	SendTS int64
}

func main() {
	// 명령줄 인자로 포트를 설정할 수 있습니다. 기본값은 32767 입니다.
	port := flag.Int("port", listenPort, "UDP listen port")
	flag.Parse()

	// net.ListenUDP를 사용하여 UDP 소켓을 열고, 클라이언트의 요청을 기다립니다.
	listenSoc := &net.UDPAddr{
		IP:   net.ParseIP(listenAddr),
		Port: *port,
	}

	udpConn, err := net.ListenUDP("udp", listenSoc)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer udpConn.Close()

	// 서버를 제어하기 위한 컨텍스트를 생성합니다.
	ctx, cancel := context.WithCancel(context.Background())
	setupSigHandlers(cancel)

	// UDP 소켓의 읽기 버퍼 크기를 설정합니다.
	if err = udpConn.SetReadBuffer(maxReadBuffer); err != nil {
		log.Fatalf("failed to set read buffer: %v", err)
	}

	log.Printf("Starting the UDP ping server")
	for {
		select {
		case <-ctx.Done():
			log.Printf("shutting down UDP server")
			return
		default:
			bytes := make([]byte, maxReadBuffer)

			if err := udpConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
				log.Printf("failed to set read deadline: %v", err)
			}

			// UDP 패킷을 읽기 위해 ReadFromUDP 호출
			// raddr: 클라이언트의 주소
			// bytes: 수신된 데이터
			// len: 데이터 길이
			len, raddr, err := udpConn.ReadFromUDP(bytes)
			if err != nil {
				log.Printf("failed to read from UDP: %v", err)
				continue
			}
			log.Printf("Received %d bytes from %s", len, raddr.String())
			if len == 0 {
				log.Printf("Received zero bytes")
				continue
			}

			// 데이터가 수신되면, 수신된 데이터를 클라이언트로
			// [:len]으로 반환하는건 뒤에 더미데이터가 있을 수 있기 때문
			n, err := udpConn.WriteToUDP(bytes[:len], raddr)
			if err != nil {
				log.Fatalf("failed to write to UDP: %v", err)
			}

			if n != len {
				log.Printf("could not send the full packet")
			}
		}
	}
}
