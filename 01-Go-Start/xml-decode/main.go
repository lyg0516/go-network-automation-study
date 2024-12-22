package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type Router struct {
	Hostname string `xml:"hostname"`
	IP       string `xml:"ip"`
	ASN      uint16 `xml:"asn"`
}

type Inventory struct {
	Routers []Router `xml:"router"`
}

func main() {
	file, err := os.Open("01-Go-Start/xml-decode/input.xml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	d := xml.NewDecoder(file)

	var inv Inventory

	err = d.Decode(&inv)
	fmt.Printf("%+v\n", inv)
}
