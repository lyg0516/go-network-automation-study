package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// 간단한 HTTP 클라이언트 프로그램으로, 사용자 입력에 따라 HTTP 서버에 요청을 보내고
// 서버의 응답을 출력하는 역할을합니다.
func main() {
	server := flag.String("server", "localhost:8080", "HTTP server URL")
	check := flag.Bool("check", false, "healthcheck flag")
	lookup := flag.String("lookup", "domain", "lookup date [mac, ip, domain]")
	flag.Parse()

	// flag.NArg()는 파싱되지 않은 입력 문자를 의미함
	// $ go run main.go -lookup=domain example.com
	// 위처럼 들어오면 check가 아닐때 남은 한개를 쿼리 파라미터로 전송
	if flag.NArg() != 1 && !(*check) {
		log.Println("must provide exactly one query argument")
		return
	}

	path := "/lookup"
	if *check {
		path = "/check"
	}

	// HTTP 서버 주소와 경로를 조합하여 URL 객체를 생성
	addr, err := url.Parse("http://" + *server + path)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP GET 요청의 쿼리 매개변수를 설정
	// params.Add(*lookup, flag.Arg(0))
	params := url.Values{}
	params.Add(*lookup, flag.Arg(0))
	// 쿼리 매개변수를 URL에 추가
	addr.RawQuery = params.Encode()

	// 생성한 URL로 GET요청을 보냄
	res, err := http.DefaultClient.Get(addr.String())
	if err != nil {
		log.Fatal(err)
	}

	// res.Body는 서버로부터 받은 응답 데이터를 읽을 수 있는 스트림
	defer res.Body.Close()

	// res.Body(응답 데이터 스트림)를 osStdout(표준 출력)으로 복사
	io.Copy(os.Stdout, res.Body)
}
