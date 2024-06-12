package main

import "fmt"

type httpResponse struct {
	version     string
	status      int
	statusMsg   string
	headers     map[string]string
	contentType string
	body        []byte
}

func NewHttpResponse(version string) *httpResponse {
	return &httpResponse{
		headers: make(map[string]string),
		version: version,
	}
}

func (res httpResponse) Parse() string {
	resStr := fmt.Sprintf("%s %d %s%s", res.version, res.status, res.statusMsg, CRLF)

	for k, v := range res.headers {
		resStr += fmt.Sprintf("%s:%s%s", k, v, CRLF)
	}
	resStr += CRLF

	resStr += string(res.body)

	return resStr
}

func (res *httpResponse) SetHeader(key, value string) {
	res.headers[key] = value
}

func (res *httpResponse) SetStatus(status int) {
	res.status = status
	switch status {
	case 200:
		res.statusMsg = "OK"
	case 404:
		res.statusMsg = "Not Found"
	case 500:
		res.statusMsg = "Internal Error"
	}
}

func (res *httpResponse) SetBody(body []byte) {
	res.body = body
}
