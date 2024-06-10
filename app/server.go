package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message: ", err.Error())
	}

	msg := string(buffer[:n])
	msg = strings.Trim(msg, " ")
	msg = strings.Replace(msg, "\n", "", -1)

	fmt.Printf("msg received (%d): %s\n", n, string(buffer[:n]))

	response := "HTTP/1.1 200 OK\r\n\r\n"

	fmt.Println("Sending response:\n", response)
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
