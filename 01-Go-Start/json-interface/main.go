package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("01-Go-Start/json-interface/input.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d := json.NewDecoder(file)

	// map[]interface를 활용해 형식을 지정하지 않고 json데이터를 받을 수 있다.
	var empty map[string]interface{}

	err = d.Decode(&empty)
	if err != nil {
		panic(err)
	}

	for _, r := range empty {
		fmt.Printf("%+v\n", r)
	}
}
