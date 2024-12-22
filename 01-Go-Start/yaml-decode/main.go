package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Router struct {
	Hostname string `yaml:"hostname"`
	IP       string `yaml:"ip"`
	ASN      uint16 `yaml:"asn"`
}

type Inventory struct {
	Routers []Router `yaml:"router"`
}

func main() {

	file, err := os.Open("01-Go-Start/yaml-decode/input.yml")
	if err != nil {
		panic(err)
	}
	d := yaml.NewDecoder(file)

	var inv Inventory

	err = d.Decode(&inv)

	fmt.Println("%+v\n", inv)
}
