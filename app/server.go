package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

const httpVer = "HTTP/1.1"

var statuses = map[int]string{
	200: "200 OK",
	404: "404 Not Found",
}

// technically the . should be only characters allowed in a URL
var echoPath = regexp.MustCompile(`^/echo/(.*)`)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// read request
	scanner := bufio.NewScanner(conn)

	// only reading the first line for now
	_ = scanner.Scan()
	line := scanner.Text()
	segments := strings.Split(line, " ")
	// method := segments[0]
	target := segments[1]
	// version := segments[2]

	var (
		statusCode  int
		respHeaders map[string]string
		echoStr     string
	)
	echoMatches := echoPath.FindStringSubmatch(target)
	if echoMatches != nil {
		echoStr = echoMatches[1]
	}

	switch {
	case echoMatches != nil:
		respHeaders = map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(echoStr)),
		}
		fallthrough
	case target == "/":
		statusCode = 200
	default:
		statusCode = 404
	}

	// respond
	var respLines []string
	// status line
	respLines = append(respLines, fmt.Sprintf("%s %s\r\n", httpVer, statuses[statusCode]))
	// headers
	for k, v := range respHeaders {
		respLines = append(respLines, fmt.Sprintf("%s: %s\r\n", k, v))
	}
	// separator between headers and body
	respLines = append(respLines, fmt.Sprint("\r\n"))
	// body
	respLines = append(respLines, echoStr)

	for _, l := range respLines {
		_, err = fmt.Fprint(conn, l)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			os.Exit(1)
		}
	}
}
