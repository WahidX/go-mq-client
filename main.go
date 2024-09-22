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

var getConn = func() net.Conn {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	return conn
}

func main() {
	// Connect to the server

	// conn, err := net.Dial("tcp", "localhost:"+port)
	// if err != nil {
	// 	log.Fatalf("Failed to connect: %v", err)
	// }
	// defer conn.Close()

	// Reader for user input
	stdInReader := bufio.NewReader(os.Stdin)

	instructions := map[string]string{
		"PING - pp":   "Ping the server",
		"PUBLISH - p": "Publish a message",
		"CONSUME - c": "Consume a message",
		"QUIT - q":    "Quit the program",
	}

	for {
		fmt.Println("\n\n============================")
		for k, ins := range instructions {
			fmt.Printf("%s\t %s\n", k, ins)
		}
		fmt.Println("============================")
		fmt.Println("Enter a command:")

		input, _ := stdInReader.ReadString('\n')

		conn := getConn()
		defer conn.Close()

		closeConn := func(err error) {
			conn.Close()
			log.Fatalf("Failed to send CONSUME command: %v", err)
		}

		switch input[:len(input)-1] {
		case "pp":
			fallthrough
		case "PING":
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
			fmt.Print("Enter topic: ")
			topic, _ := stdInReader.ReadString('\n')
			fmt.Print("Enter binary message to send: ")
			msg, _ := stdInReader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			// starting to send the inputs to the MQ server
			// PUBLISH command (1 byte)
			if _, err := conn.Write([]byte{1}); err != nil {
				closeConn(err)
			}

			// Send the topic
			if _, err := conn.Write([]byte(topic)); err != nil {
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
			topic, _ := stdInReader.ReadString('\n')

			if _, err := conn.Write([]byte(topic)); err != nil {
				closeConn(err)
			}

			// Wait and keep reading incoming data
			msg := make([]byte, 1024)

			for {
				n, err := conn.Read(msg)
				fmt.Println("n: ", n)
				if err != nil {
					fmt.Println("Err in reading message: ", err)
					break
				}

				fmt.Println("incoming data: ", string(msg))
				time.Sleep(3 * time.Second)
			}

		case "q":
			fallthrough
		case "QUIT":
			fmt.Println("See you soon...")
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
