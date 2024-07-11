package main

import (
	"fmt"
	"net"
	"os"
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

	status := "200 OK"

	conn.Write([]byte(fmt.Sprintf("%s %s\r\n\r\n", httpVer, status)))
}
