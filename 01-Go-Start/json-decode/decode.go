package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Router struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	ASN      uint16 `json:"asn"`
}

type Inventory struct {
	Routers []Router `json:"router"`
}

func main() {
	file, err := os.Open("01-Go-Start/input.json")
	// 에러를 처리한다.
	if err != nil {
		panic("failed to open file")
	}
	defer file.Close()

	// io.Reader 타입에서 데이터 추출
	d := json.NewDecoder(file)

	var inv Inventory

	err = d.Decode(&inv)

	// %v는 구조체 서식이고 +하면 id를 포함해서 출력
	fmt.Printf("%+v\n", inv)
}
