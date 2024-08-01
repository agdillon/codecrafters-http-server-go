package main

import (
	"bufio"
	"flag"
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
var filesPath = regexp.MustCompile(`^/files/(.*)`)

func main() {
	// handle command line flag for files directory
	directory := flag.String("directory", "/tmp/", "absolute path to directory to read files for file endpoint")
	flag.Parse()
	directoryStr := *directory
	if string(directoryStr[len(*directory)-1]) != "/" {
		directoryStr = directoryStr + "/"
	}

	// listen for connections
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn, directoryStr)
	}
}

func handleRequest(conn net.Conn, directoryStr string) {
	// read request
	scanner := bufio.NewScanner(conn)

	// read request line
	_ = scanner.Scan()
	line := scanner.Text()
	segments := strings.Split(line, " ")
	// method := segments[0]
	target := segments[1]
	// version := segments[2]

	// read headers
	reqHeaders := map[string]string{}
	for {
		more := scanner.Scan()
		line := scanner.Text()
		kv := strings.Split(line, ":")
		if len(kv) >= 2 {
			headerName := strings.ToLower(strings.TrimSpace(kv[0]))
			reqHeaders[headerName] = strings.TrimSpace(kv[1])
		}

		if line == "" || !more {
			// handle errors here
			break
		}
	}

	var (
		echoStr     string
		filesStr    string
		respBody    string
		respHeaders map[string]string
		statusCode  int
	)
	// match against /echo/{str} and extract parameter
	echoMatches := echoPath.FindStringSubmatch(target)
	if len(echoMatches) >= 2 {
		echoStr = echoMatches[1]
	}

	// match against /files/{filename} and extract parameter
	// TODO - combine this with echo
	filesMatches := filesPath.FindStringSubmatch(target)
	if len(filesMatches) >= 2 {
		filesStr = filesMatches[1]
	}

	switch {
	case echoMatches != nil:
		respBody = echoStr
		respHeaders = map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(respBody)),
		}
		statusCode = 200
	case filesMatches != nil:
		dat, err := os.ReadFile(fmt.Sprintf("/%s%s", directoryStr, filesStr))
		if err != nil {
			statusCode = 404
		} else {
			respBody = string(dat)
			respHeaders = map[string]string{
				"Content-Type":   "application/octet-stream",
				"Content-Length": fmt.Sprintf("%d", len(respBody)),
			}
			statusCode = 200
		}
	case target == "/user-agent":
		respBody = reqHeaders["user-agent"]
		respHeaders = map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(respBody)),
		}
		statusCode = 200
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
	respLines = append(respLines, "\r\n")
	// body
	respLines = append(respLines, respBody)

	for _, l := range respLines {
		_, err := fmt.Fprint(conn, l)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			os.Exit(1)
		}
	}
}
