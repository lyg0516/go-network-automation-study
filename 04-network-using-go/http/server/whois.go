package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// 이 코드는 whois 프로토콜을 통해 도메인 정보를 조회하는 기능을 제공하는 프로그램입니다
const (
	// 기본 WHOIS 서버
	whoisIANA = "whois.iana.org"

	// WHOIS 서비스가 사용하는 기본 포트
	whoisPort = 43
)

// WHOIS 서버에 쿼리를 보내고 결과를 반환합니다.
func whoisLookup(query, server string) (*bytes.Buffer, error) {
	log.Printf("whoisLookup: %s@%s", query, server)

	server = fmt.Sprintf("%s:%d", server, whoisPort)

	// 주어진 서버 이름(server)를 TCP 주소로 변환합니다.
	rAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr failed: %s", err)
	}

	// WHOIS 서버에 TCP 연결을 생성합니다. IPv4를 사용하며, 연결 실패 시 에러를 반환합니다.
	conn, err := net.DialTCP("tcp4", nil, rAddr)
	if err != nil {
		return nil, fmt.Errorf("DialTCP failed: %s", err)
	}
	defer conn.Close()

	// all queries must end with CRLF
	// conn.Write로 데이터를 보내고 이후 conn
	query += "\r\n"
	_, err = conn.Write([]byte(query))
	if err != nil {
		return nil, fmt.Errorf("Write failed: %s", err)
	}

	var response bytes.Buffer

	_, err = io.Copy(&response, conn)
	if err != nil {
		return nil, fmt.Errorf("Read failed: %s", err)
	}

	return &response, nil
}

// findRefer는 WHOIS 응답 데이터를 분석하여 refer라는 키워드가 포함된 줄에서 "참조 WHOIS 서버"를 찾는
// 기능을 수행합니다. 이 정보를 사용해 추가적으로 다른 WHOIS 서버를 조회할 수 있도록 돕습니다.
// whois 응답
/*
	Domain Name: example.com
	Registrar: Example Registrar
	Updated Date: 2024-01-01T00:00:00Z
	Creation Date: 2020-01-01T00:00:00Z
	refer: whois.example-registrar.com

# The above server will provide detailed information.

*/
func findRefer(input *bytes.Buffer) (string, bool) {
	lineScanner := bufio.NewScanner(input)
	for lineScanner.Scan() {
		if len(lineScanner.Bytes()) < 1 {
			continue
		}
		if lineScanner.Bytes()[0] == '#' {
			continue
		}

		line := lineScanner.Text()

		if strings.Contains(line, "refer: ") {
			parts := strings.Fields(line)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return "", false
			}
			return parts[1], true
		}
	}
	if err := lineScanner.Err(); err != nil {
		return "", false
	}

	return "", false

}
