package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// 이 코드는 간단한 HTTP 서버를 구현한 프로그램으로, 특정 API 경로(/lookup, /check)
// 에 따라 다양한 네트워크 관련 작업(WHOIS 조회, MAC 주소 조회 등)을 처리합니다.

// 1) /lookup API
// 클라이언트가 ip, mac, 또는 domain 파라미터를 포함한 GET 요청을 보내면 해당 값에 대한 작업을 수행합니다.
// ip 또는 domain: WHOIS 정보를 조회
// mac: MAC 주소의 OUI(Organization Unique Identifier)를 기반으로 벤더정보를 반환
// 2) /check API
// 단순 상태 확인 API
// 클라이언트가 /check 경로로 요청을 보내면 "OK"라는 응답을 반환

// WHOIS 정보를 조회하는 함수
func getWhois(input []string) string {

	// input이 하나인지 확인
	if len(input) != 1 {
		return fmt.Sprintf("incorrect query %v", input)
	}

	var res string
	query := input[0]
	whoisServer := whoisIANA
	for {
		response, err := whoisLookup(query, whoisServer)
		if err != nil {
			log.Fatalf("lookup failes: %s", err)
		}

		res = response.String()

		refer, found := findRefer(response)
		if found {
			whoisServer = refer
			continue
		}

		break
	}
	return res
}

func getMAC(input []string) string {
	if len(input) != 1 {
		return fmt.Sprintf("incorrect query %v", input)
	}

	// MAC 주소를 파싱하고 OUI(상위 3바이트)를 추출
	// 이후 OUI에 대해서 macDB조회
	mac, err := net.ParseMAC(input[0])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	oui := mac[:3].String()
	oui = strings.ToUpper(oui)

	res, ok := macDB[oui]
	if !ok {
		return fmt.Sprintf("result not found\n")
	}
	return res
}

func lookup(w http.ResponseWriter, req *http.Request) {

	log.Printf("Incoming %+v", req.URL.Query())
	var response string

	for k, v := range req.URL.Query() {
		switch k {
		case "ip":
			response = getWhois(v)
		case "mac":
			response = getMAC(v)
		case "domain":
			response = getWhois(v)
		default:
			response = fmt.Sprintf("query %q not recognized\n", k)
		}
	}
	fmt.Fprintf(w, response)
}

func check(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func main() {
	http.HandleFunc("/lookup", lookup)

	http.HandleFunc("/check", check)

	log.Println("Starting web server at 0.0.0.0:8080")

	//http.Server 객체를 생성하며, Addr 필드에 서버가 바인딩될 IP 주소와 포트 설정
	// "0.0.0.0": 모든 네트워크 인터페이스에서 연결을 수락
	// 8080: 서버가 사용할 포트 번호
	srv := http.Server{Addr: "0.0.0.0:8080"}
	log.Fatal(srv.ListenAndServe())
}
