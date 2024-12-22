package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type Router struct {
	Hostname string `json:"hostname" xml:"hostname"`
	IP       string `json:"ip" xml:"ip"`
	ASN      uint16 `json:"asn" xml:"asn"`
}

type Inventory struct {
	Routers []Router `json:"router" xml:"router"`
}

func main() {
	file, err := os.Open("01-Go-Start/xml-encode/input.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d := json.NewDecoder(file)

	var inv Inventory
	err = d.Decode(&inv)

	var dest strings.Builder
	fmt.Printf("%+v\n", inv)
	e := xml.NewEncoder(&dest)
	err = e.Encode(&inv)

	fmt.Printf("%+v\n", dest.String())
}
