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

type httpResquest struct {
	method      string
	path        string
	httpVersion string
	host        string
	headers     map[string]string
	body        string
}

func newHttpRequest(rStr string) *httpResquest {
	requestLine, rStr, found := strings.Cut(rStr, "\r\n")
	if !found {
		return nil
	}

	var r httpResquest

	rLine := strings.Split(requestLine, " ")
	r.method = rLine[0]
	r.path = rLine[1]
	r.httpVersion = rLine[2]

	return &r
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message: ", err.Error())
	}

	msg := string(buffer[:n])
	r := newHttpRequest(msg)
	fmt.Println(r)

	var response string
	if strings.Compare("/", r.path) == 0 {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	fmt.Println("Sending response:\n", response)
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
