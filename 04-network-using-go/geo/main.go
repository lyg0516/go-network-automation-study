package main

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
)

/*
MaxMind의 GeoIP2 데이터베이스를 활용하여 IP 주소의 지리적 정보(도시, 국가, 시간대 등)를 검색하는 프로그램입니다.
*/
func main() {
	db, err := geoip2.Open("04-network-using-go/geo/GeoIP2-City-Test.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	IPs := []string{
		"216.160.83.57",
		"2001:270::f00d",
		"81.2.69.143",
		"2001:2a0::cafe",
	}

	fmt.Println("Find information for each prefix:")
	for _, prefix := range IPs {
		ip := net.ParseIP(prefix)
		record, err := db.City(ip)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nAddress: %v\n", prefix)
		fmt.Printf("City name: %v\n", record.City.Names["en"])
		fmt.Printf("Country name: %v\n", record.Country.Names["en"])
		fmt.Printf("ISO country code: %v\n", record.Country.IsoCode)
		fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
		fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
		fmt.Println("--------------------------------------------")
	}

}
