package server

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
)

// Wireshark 프로젝트의 MAC 주소 데이터베이스를 다운로드하고, 이를 파싱하여
// macDB라는 전역 맵에 저장하는 기능
const wiresharkDB = "https://gitlab.com/wireshark/wireshark.git"

var macDB map[string]string

// init 함수는 프로그램 실행 시 자동으로 호출되며, macDB를 초기화합니다.
func init() {
	var err error
	macDB, err = download()
	if err != nil {
		log.Fatalf("Failed to download mac database: %v", err)
	}
	log.Printf("macDB initialized")
}

func parse(db io.Reader, out map[string]string) map[string]string {
	lineScanner := bufio.NewScanner(db)
	for lineScanner.Scan() {
		if len(lineScanner.Bytes()) < 1 {
			continue
		}
		if lineScanner.Bytes()[0] == '#' {
			continue
		}

		parts := strings.Split(lineScanner.Text(), "\t")

		if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
			continue
		}

		out[parts[0]] = parts[2]
	}

	if err := lineScanner.Err(); err != nil {
		return out
	}
	return out
}

func download() (map[string]string, error) {
	result := make(map[string]string)

	resp, err := http.Get(wiresharkDB)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	parse(resp.Body, result)

	return result, nil
}
