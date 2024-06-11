package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTP_VERSION = "HTTP/1.1"
	CRLF         = "\r\n"
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

type httpResponse struct {
	version     string
	status      int
	statusMsg   string
	headers     map[string]string
	contentType string
	body        []byte
}

func newHttpResponse(versin string, status int, statusMsg string, headers map[string]string, body []byte) (res string, err error) {
	headLine := fmt.Sprintf("%s %d %s%s", versin, status, statusMsg, CRLF)
	headersStr := ""
	for k, v := range headers {
		headersStr += fmt.Sprintf("%s: %s%s", k, v, CRLF)
	}
	headersStr += fmt.Sprintf("Content-Length: %d%s%s", len(body), CRLF, CRLF)

	return headLine + headersStr + string(body), nil
}

func (res httpResponse) parse() string {
	resStr := fmt.Sprintf("%s %d %s%s", res.version, res.status, res.statusMsg, CRLF)

	for k, v := range res.headers {
		resStr += fmt.Sprintf("%s:%s%s", k, v, CRLF)
	}
	resStr += fmt.Sprintf("Content-Length:%d%s", len(res.body), CRLF)
	resStr += fmt.Sprintf("Content-Type:%s%s", res.contentType, CRLF)
	resStr += CRLF

	resStr += string(res.body)

	return resStr
}

type httpResquest struct {
	method      string
	path        string
	httpVersion string
	host        string
	headers     map[string]string
	body        []byte
}

func parseHeadLine(line []byte) (method, path, version string, err error) {
	partStr := strings.ReplaceAll(string(line), "  ", " ")
	parts := strings.Split(partStr, " ")

	if len(parts) != 3 {
		err = fmt.Errorf("invalid request line")
		return
	}

	method = parts[0]
	path = parts[1]
	version = parts[2]

	return
}

func parserHeader(line []byte) (key, value string, err error) {
	parts := strings.SplitN(string(line), ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid header format")
		return
	}
	return parts[0], parts[1], nil
}

func parseRequest(rBytes []byte) (*httpResquest, error) {
	lines := bytes.Split(rBytes, []byte(CRLF))
	method, path, version, err := parseHeadLine(lines[0])
	if err != nil {
		return nil, err
	}

	request := httpResquest{
		method:      method,
		path:        path,
		httpVersion: version,
		headers:     map[string]string{},
		body:        make([]byte, 1024),
	}

	var isParsingHeaders bool

	for _, line := range lines[1:] {
		switch {
		case !isParsingHeaders && bytes.Contains(line, []byte(":")):
			key, value, err := parserHeader(line)
			if err != nil {
				return nil, err
			}

			request.headers[strings.TrimSpace(key)] = strings.TrimSpace(value)

		case bytes.Equal(line, []byte(CRLF)):
			isParsingHeaders = true

		default:
			request.body = append(request.body, line...)
			request.body = append(request.body, []byte(CRLF)...)
		}
	}

	if len(request.body) > 0 {
		request.body = request.body[:len(request.body)-len([]byte(CRLF))]
	}

	return &request, nil
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message: ", err.Error())
	}

	response := httpResponse{
		version:     HTTP_VERSION,
		contentType: "text/plain",
	}

	r, err := parseRequest(buffer)
	if err != nil {
		fmt.Println("Erro parsing request: ", err.Error())
		response = httpResponse{
			status:    500,
			statusMsg: "Internal Error",
		}
	} else {
		fmt.Printf("path: |%s| %b\n", (*r).path, strings.Compare("/", (*r).path))
		for k, v := range (*r).headers {
			fmt.Printf("%s: %s\n", k, v)
		}
		if strings.Compare("/", r.path) == 0 {
			response.status = 200
			response.statusMsg = "OK"
		} else if strings.Compare("/user-agent", (*r).path) == 0 {
			userAgent, userAgentIsPresent := (*r).headers["User-Agent"]

			if userAgentIsPresent {
				response.status = 200
				response.statusMsg = "OK"
				response.body = []byte(userAgent)
			} else {
				response.status = 400
			}

		} else if strings.HasPrefix((*r).path, "/echo/") {
			bodyContent := strings.Replace((*r).path, "/echo/", "", 1)
			response.status = 200
			response.statusMsg = "OK"
			response.body = []byte(bodyContent)
		} else {
			response.status = 404
			response.statusMsg = "Not Found"
		}
	}

	fmt.Println("Sending response:\n", response)
	_, err = conn.Write([]byte(response.parse()))
	if err != nil {
		fmt.Println("Error sending response: ", err.Error())
	}
}
