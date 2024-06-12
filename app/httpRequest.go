package main

import (
	"bytes"
	"fmt"
	"strings"
)

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

func ParseRequest(rBytes []byte) (*httpResquest, error) {
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
