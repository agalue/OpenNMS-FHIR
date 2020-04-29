package main

import (
	"flag"
	"log"
	"net"

	"sample-generator/api/telemetry_bis"

	"github.com/golang/protobuf/proto"
)

var totalSteps = 0

func main() {
	source := flag.String("source", "localhost:4444", "Source for the UDP Telemetry Data")
	flag.Parse()

	log.Printf("Listening for Telemetry Data on %s", *source)
	s, err := net.ResolveUDPAddr("udp4", *source)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ERROR: %s", err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, addr)
		telemetry := &telemetry_bis.Telemetry{}
		err = proto.Unmarshal(buf[:n], telemetry)
		if err != nil {
			log.Printf("ERROR: %s", err)
		} else {
			log.Printf("Telemetry Data: %s", telemetry.String())
		}
	}
}
