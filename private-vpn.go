package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const (
	serverAddr = "220.149.231.241:9091"
	vpnAddr    = "10.120.0.101:1194"
)

func main() {
	// Connect to TCP server
	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer serverConn.Close()
	fmt.Println("Connected to server on", serverAddr)

	// Connect to VPN (now TCP instead of UDP)
	vpnConn, err := net.Dial("tcp", vpnAddr)
	if err != nil {
		log.Fatal("Error connecting to VPN:", err)
	}
	defer vpnConn.Close()
	fmt.Println("Connected to VPN at", vpnAddr)

	// Start goroutine to handle server to VPN traffic
	go forwardData(serverConn, vpnConn, "server", "VPN")

	// Handle VPN to server traffic
	forwardData(vpnConn, serverConn, "VPN", "server")
}

func forwardData(src, dst net.Conn, srcName, dstName string) {
	buffer := make([]byte, 1024)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %v\n", srcName, err)
			}
			return
		}

		payload := buffer[:n]
		//fmt.Printf("Received from %s: %s\n", srcName, payload)

		_, err = dst.Write(payload)
		if err != nil {
			log.Printf("Error sending to %s: %v\n", dstName, err)
			continue
		}
		//fmt.Printf("Forwarded to %s: %s\n", dstName, payload)
	}
}
