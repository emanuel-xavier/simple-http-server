package main

import (
	"bytes"
	"fmt"
	"strings"
)

type httpRequest struct {
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

func parseHeader(line []byte) (key, value string, err error) {
	parts := strings.SplitN(string(line), ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid header format")
		return
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func ParseRequest(rBytes []byte) (*httpRequest, error) {
	lines := bytes.Split(rBytes, []byte(CRLF))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	method, path, version, err := parseHeadLine(lines[0])
	if err != nil {
		return nil, err
	}

	request := httpRequest{
		method:      method,
		path:        path,
		httpVersion: version,
		headers:     make(map[string]string),
	}

	var i int
	for i = 1; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			break
		}

		key, value, err := parseHeader(line)
		if err != nil {
			return nil, err
		}

		request.headers[key] = value
	}

	if i < len(lines)-1 {
		request.body = bytes.Join(lines[i+1:], []byte(CRLF))
	}

	return &request, nil
}
