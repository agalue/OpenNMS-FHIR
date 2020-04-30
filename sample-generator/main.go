package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"sample-generator/api/telemetry_bis"

	"github.com/golang/protobuf/proto"
)

var totalSteps = 0

func main() {
	target := flag.String("target", "localhost:4444", "Destination for the UDP Telemetry Data")
	frequency := flag.Duration("frequency", 30*time.Second, "Frequency of packet generation")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "generator"
	}

	rand.Seed(time.Now().Unix())

	src, err := net.ResolveUDPAddr("udp4", *target)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp4", nil, src)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("Sending UDP Telemetry packets to %s every %s", *target, (*frequency).String())
	for {
		t := createTelemetryData(hostname)
		if err := sendData(conn, t); err != nil {
			log.Printf("ERROR: %s", err)
		}
		time.Sleep(*frequency)
	}
}

func createTelemetryData(hostname string) *telemetry_bis.Telemetry {
	totalSteps += rand.Intn(100)
	ts := uint64(time.Now().Unix())
	return &telemetry_bis.Telemetry{
		NodeId: &telemetry_bis.Telemetry_NodeIdStr{
			NodeIdStr: hostname,
		},
		MsgTimestamp: ts,
		DataGpbkv: []*telemetry_bis.TelemetryField{
			{
				Name:      "heartRate", // Gauge
				Timestamp: ts,
				ValueByType: &telemetry_bis.TelemetryField_Uint32Value{
					Uint32Value: uint32(60 + rand.Intn(100)),
				},
			},
			{
				Name:      "stepCount", // Counter
				Timestamp: ts,
				ValueByType: &telemetry_bis.TelemetryField_Uint32Value{
					Uint32Value: uint32(totalSteps),
				},
			},
		},
	}
}

func sendData(conn net.Conn, telemetry *telemetry_bis.Telemetry) error {
	log.Printf("Sending %s\n", telemetry.String())
	msg, err := proto.Marshal(telemetry)
	if err != nil {
		return err
	}
	n, err := conn.Write(msg)
	if err != nil {
		return err
	}
	log.Printf("Sent %d bytes", n)
	return nil
}
