package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	FILES_PATH string
)

func init() {
	flag.StringVar(&FILES_PATH, "directory", "/tmp", "Files directory")
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	flag.Parse()
	fmt.Println("FILEPATH ", FILES_PATH)

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

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message: ", err.Error())
	}

	response := NewHttpResponse(HTTP_VERSION)

	r, err := ParseRequest(buffer)
	if err != nil {
		fmt.Println("Erro parsing request: ", err.Error())
		response.SetStatus(STATUS_INTERNAL_ERROR)
	} else {

		fmt.Printf("path: %s\n", r.path)

		switch {
		case strings.Compare("/", r.path) == 0:
			response.SetStatus(STATUS_OK)

		case strings.HasPrefix(r.path, "/files/"):
			fileName, found := strings.CutPrefix(r.path, "/files/")
			filePath := fmt.Sprintf("%s%s", FILES_PATH, fileName)
			response.SetStatus(STATUS_NOT_FOUND)

			if file, err := os.Open(filePath); err == nil && found {
				fileContent, err := io.ReadAll(file)
				if err == nil {
					fmt.Println("body content:\n", string(fileContent))
					response.SetStatus(STATUS_OK)
					response.SetHeader("Content-Type", "application/octet-stream")
					response.SetHeader("Content-Length", strconv.Itoa(len(fileContent)))
					response.SetBody(fileContent)
				}
				file.Close()
			}

		case strings.Compare("/user-agent", r.path) == 0:
			userAgent, userAgentIsPresent := r.headers["User-Agent"]
			if userAgentIsPresent {
				bodyBytes := []byte(userAgent)
				response.SetStatus(STATUS_OK)
				response.SetBody(bodyBytes)
				response.SetHeader("Content-Length", strconv.Itoa(len(bodyBytes)))
				response.SetHeader("Content-Type", "text/plain")
			} else {
				response.SetStatus(STATUS_BAD_REQUEST)
			}

		case strings.HasPrefix(r.path, "/echo/"):
			bodyContent := strings.Replace(r.path, "/echo/", "", 1)
			bodyBytes := []byte(bodyContent)
			response.SetStatus(STATUS_OK)
			response.SetBody(bodyBytes)
			response.SetHeader("Content-Length", strconv.Itoa(len(bodyBytes)))
			response.SetHeader("Content-Type", "text/plain")

		default:
			response.SetStatus(STATUS_NOT_FOUND)
		}

	}

	_, err = conn.Write([]byte(response.Parse()))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
