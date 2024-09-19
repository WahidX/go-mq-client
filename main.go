package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var port = "4000"

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Reader for user input
	reader := bufio.NewReader(os.Stdin)

	closeConn := func(err error) {
		conn.Close()
		log.Fatalf("Failed to send CONSUME command: %v", err)
	}

	for {
		fmt.Print("Enter command (PUBLISH or CONSUME) (p/c): ")
		input, _ := reader.ReadString('\n')

		switch input[:len(input)-1] {
		case "ping":
			t := time.Now()
			fmt.Println("Pinging...")

			if _, err := conn.Write([]byte{0}); err != nil { // ping command - 0
				closeConn(err)
			}

			var pingResBin uint32
			err := binary.Read(conn, binary.BigEndian, &pingResBin)
			if err != nil {
				closeConn(err)
			}
			if pingResBin == 1 {
				fmt.Println("Pong", time.Since(t))
			} else {
				fmt.Println("Failed to ping", time.Since(t))
			}

		case "p":
			fallthrough
		case "PUBLISH":
			fmt.Print("Enter binary message to send: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			// starting to send the inputs to the MQ server
			// PUBLISH command (1 byte)
			if _, err := conn.Write([]byte{1}); err != nil {
				closeConn(err)
			}

			// Send message length (4 bytes)
			if err := binary.Write(conn, binary.BigEndian, uint32(len(msg))); err != nil {
				closeConn(err)
			}

			// Send the message
			if _, err := conn.Write([]byte(msg)); err != nil {
				closeConn(err)
			}

			// Read the server's acknowledgment
			var response []byte
			_, _ = bufio.NewReader(conn).Read(response)
			fmt.Printf("Server Response: %s\n", response)

		case "c":
			fallthrough
		case "CONSUME":
			// Send CONSUME command (1 byte)
			if _, err := conn.Write([]byte{2}); err != nil {
				closeConn(err)
			}

			fmt.Print("Enter topic to consume: ")
			topic, _ := reader.ReadString('\n')

			if _, err = conn.Write([]byte(topic)); err != nil {
				closeConn(err)
			}

			// Read the length of the binary message (4 bytes)
			// Read the binary message

		default:
			fmt.Println("Unknown command")
		}
	}
}
