package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const (
	incomingAddr = ":9092"
	outgoingAddr = ":9091"
)

func main() {
	// Start TCP listener for incoming connections (previously UDP)
	incomingListener, err := net.Listen("tcp", incomingAddr)
	if err != nil {
		log.Fatal("Error starting incoming TCP server:", err)
	}
	defer incomingListener.Close()
	fmt.Println("Incoming TCP server listening on", incomingAddr)

	// Start TCP listener for outgoing connections
	outgoingListener, err := net.Listen("tcp", outgoingAddr)
	if err != nil {
		log.Fatal("Error starting outgoing TCP server:", err)
	}
	defer outgoingListener.Close()
	fmt.Println("Outgoing TCP server listening on", outgoingAddr)

	var outgoingConn net.Conn
	var outgoingConnMutex sync.Mutex

	// Handle outgoing connection (to client's :1194)
	go func() {
		for {
			conn, err := outgoingListener.Accept()
			if err != nil {
				log.Println("Error accepting outgoing TCP connection:", err)
				continue
			}
			outgoingConnMutex.Lock()
			if outgoingConn != nil {
				outgoingConn.Close()
			}
			outgoingConn = conn
			outgoingConnMutex.Unlock()
			fmt.Println("New outgoing TCP client connected from", conn.RemoteAddr())
		}
	}()

	// Handle incoming connections (from original sender)
	for {
		incomingConn, err := incomingListener.Accept()
		if err != nil {
			log.Println("Error accepting incoming TCP connection:", err)
			continue
		}
		fmt.Println("New incoming TCP client connected from", incomingConn.RemoteAddr())

		go handleConnection(incomingConn, &outgoingConn, &outgoingConnMutex)
	}
}

func handleConnection(incomingConn net.Conn, outgoingConn *net.Conn, outgoingConnMutex *sync.Mutex) {
	defer incomingConn.Close()

	// Forward data from incoming to outgoing
	go forwardData(incomingConn, outgoingConn, outgoingConnMutex, "incoming", "outgoing")

	// Forward data from outgoing to incoming
	forwardData(*outgoingConn, &incomingConn, nil, "outgoing", "incoming")
}

func forwardData(src net.Conn, dst *net.Conn, dstMutex *sync.Mutex, srcName, dstName string) {
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

		if dstMutex != nil {
			dstMutex.Lock()
		}
		if *dst != nil {
			_, err = (*dst).Write(payload)
			if err != nil {
				log.Printf("Error forwarding to %s: %v\n", dstName, err)
				(*dst).Close()
				*dst = nil
				break
			} else {
				//fmt.Printf("Payload forwarded to %s\n", dstName)
			}
		} else {
			fmt.Printf("No %s connection, cannot forward payload\n", dstName)
		}
		if dstMutex != nil {
			dstMutex.Unlock()
		}
	}
}
