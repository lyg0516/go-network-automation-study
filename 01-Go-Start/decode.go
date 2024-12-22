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

	d := json.NewDecoder(file)

	var inv Inventory

	err = d.Decode(&inv)

	fmt.Printf("%+v\n", inv)
}
