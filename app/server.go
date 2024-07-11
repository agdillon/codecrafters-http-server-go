package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	const httpVer = "HTTP/1.1"

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

	var statuses = map[int]string{
		200: "200 OK",
		404: "404 Not Found",
	}

	var req = make([]byte, 500)
	conn.Read(req)

	reqStr := string(req)
	fmt.Printf("hello, %s", reqStr)
	lines := strings.Split(reqStr, "\r\n")

	reqLine := lines[0]
	segments := strings.Split(reqLine, " ")
	// method := segments[0]
	target := segments[1]
	// version := segments[2]

	var statusCode int
	switch target {
	case "/":
		statusCode = 200
	default:
		statusCode = 404
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%s %s\r\n\r\n", httpVer, statuses[statusCode])))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}
}
