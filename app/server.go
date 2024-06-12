package main

import (
	"flag"
	"fmt"
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

	os.MkdirAll(FILES_PATH, os.ModePerm)

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

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message: ", err.Error())
	}

	response := NewHttpResponse(HTTP_VERSION)

	r, err := ParseRequest(buffer[:n])
	if err != nil {
		fmt.Println("Erro parsing request: ", err.Error())
		response.SetStatus(STATUS_INTERNAL_ERROR)
	} else {

		fmt.Printf("%s %s\n", r.method, r.path)

		switch {
		case strings.Compare("/", r.path) == 0:
			response.SetStatus(STATUS_OK)

		case r.method == "POST" && strings.HasPrefix(r.path, "/files/"):
			fileName, found := strings.CutPrefix(r.path, "/files/")
			if !found {
				response.SetStatus(STATUS_NOT_FOUND)
			} else {
				filePath := fmt.Sprintf("%s%s", FILES_PATH, fileName)

				err := os.WriteFile(filePath, r.body, 0644)
				if err != nil {
					response.SetStatus(STATUS_INTERNAL_ERROR)
					fmt.Println("Failed to write into file: ", err.Error())
				}
				response.SetStatus(STATUS_CREATED)
				response.SetHeader("Content-Type", "text/plain")
			}

		case r.method == "GET" && strings.HasPrefix(r.path, "/files/"):
			fileName, found := strings.CutPrefix(r.path, "/files/")
			filePath := fmt.Sprintf("%s%s", FILES_PATH, fileName)
			if !found {
				response.SetStatus(STATUS_NOT_FOUND)
			} else {
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Println("Failed to read file: ", err.Error())
					response.SetStatus(STATUS_NOT_FOUND)
				} else {
					contentLength := strconv.Itoa(len(fileContent))
					response.SetStatus(STATUS_OK)
					response.SetHeader("Content-Type", "application/octet-stream")
					response.SetHeader("Content-Length", contentLength)
					response.SetBody(fileContent)
				}
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

	fmt.Println("-----------------------------")
	fmt.Println(response.Parse())
	fmt.Println("-----------------------------")
	_, err = conn.Write([]byte(response.Parse()))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
