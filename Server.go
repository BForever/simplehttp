package main

import (
	"net"
	"fmt"
	"bufio"
	"bytes"
	"strings"
	"log"
	"strconv"
	"os"
	"io/ioutil"
	"time"
)

/*
 Use 127.0.0.1:8080 to access the server
 */
var root string
var root_ex string
// var head string = "HTTP/1.1 200 OK\nAccept-Ranges: bytes\n"
var head string = "HTTP/1.1 200 OK\n"
var notfound string = "HTTP/1.1 404 Not Found\n"

func main() {
	// Get Web Root
	if len(os.Args) < 3 {
		log.Fatal("Usage: Server [RealRoot] [PseudoRoot]")
	}
	root = os.Args[1]
	root_ex = os.Args[2]
	print("RealRoot directory at ")
	println(root)
	print("PseudoRoot directory at ")
	println(root_ex)

	// Start listening
	ln, err := net.Listen("tcp", ":8080")

	// Handle error
	if err != nil {
		log.Fatal(err)
	}
	// Wait for connections or Exit command
	chan_quit := make(chan int)
	chan_listened := make(chan net.Conn)
	go waitquit(chan_quit)
	go listen(chan_listened, ln)
	for {
		select {
		case <-chan_quit:
			goto exit
		case newc := <-chan_listened:
			go connected(newc)
		}
	}
exit:
}
func connected(conn net.Conn) {
	// Read Buffer
	time.Sleep(500000000)
	buf := bufio.NewReader(conn)
	// Write Buffer
	var writebuf bytes.Buffer
	// Wait for data to come
	_, err := buf.Peek(1)
	if err != nil {
		log.Fatal(err)
	}

	// Handle request
	handleRequest(buf, &writebuf)
	// Send respond
	if writebuf.Len() > 0 {
		print(writebuf.String())
		conn.Write(writebuf.Bytes())
	}
	// Disconnect
	conn.Close()
}
func handleRequest(inbuf *bufio.Reader, outbuf *bytes.Buffer) {
	if inbuf.Buffered() > 0 {
		// Peek the request type
		data, _, _ := inbuf.ReadLine()
		println(string(data))
		line := strings.Split(string(data), " ")[0]
		path := strings.Split(string(data), " ")[1]
		// Separate two types of request
		if strings.Contains(line, "GET") {
			handleGetReq(inbuf, path, outbuf)
		} else if strings.Contains(line, "POST") {
			handlePostReq(inbuf, path, outbuf)
		} else {
			// Request not supported
		}
	}
}
func parsePath(path string) (string, bool) {
	if path == root_ex {
		// Default root
		path = root + "test.html"
	} else {
		if strings.HasPrefix(path, root_ex) {
			// Mapping path
			path = strings.TrimLeft(path, root_ex)
			path = root + path
		} else {
			path = root + strings.TrimLeft(path, "/")
		}
	}
	// Check whether the file exists
	_, err := os.Stat(path)
	if err != nil {
		// 404 Not Found
		return "", false
	}
	return path, true
}

func clearInputBuf(inbuf *bufio.Reader) {
	for inbuf.Buffered() > 0 {
		line, _, _ := inbuf.ReadLine()
		str := string(line)
		println(str)
		if len(str) == 0 {
			// line, _, _ = inbuf.ReadLine()
			println("buf clear, new data")
			return
		}
	}
	println("buf cleared, no data")
}
func handleGetReq(inbuf *bufio.Reader, path string, outbuf *bytes.Buffer) {
	// Tmp buf to store response head
	var tmpbuf bytes.Buffer
	// Get real path
	path, exist := parsePath(path)
	// Check file
	if (!exist) {
		// File not exists, 404
		clearInputBuf(inbuf)
		NotFound(outbuf)
		return
	}
	// File exists, clear input buf
	println("parsed path: " + path)
	clearInputBuf(inbuf)

	// 200 OK head
	writeOKHead(&tmpbuf)
	// Select file type
	vec := strings.Split(path, ".")
	suffix := vec[len(vec)-1]
	switch {
	case suffix == "html":
		{
			tmpbuf.Write([]byte("Content-Type: text/html"))
			tmpbuf.Write([]byte("; charset=utf-8\n"))
			break
		}
	case suffix == "txt":
		{
			tmpbuf.Write([]byte("Content-Type: text/txt"))
			tmpbuf.Write([]byte("; charset=utf-8\n"))
			break
		}
	case suffix == "jpg":
		{
			tmpbuf.Write([]byte("Content-Type: image/jpg"))
			break
		}
	default:
		{
			NotFound(outbuf)
			goto finish
		}
	}
	// Write whole head
	outbuf.Write(tmpbuf.Bytes())
	// Send file
	writeFileBytes(path, outbuf)
finish:
}

func writeFileBytes(filename string, outbuf *bytes.Buffer) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	outbuf.Write([]byte("Content-Length: " + strconv.Itoa(len(data)) + "\n\n"))
	outbuf.Write(data)
}
func handlePostReq(inbuf *bufio.Reader, path string, outbuf *bytes.Buffer) {
	// Check post
	if path != root_ex+"dopost" {
		// Post not exists, 404
		clearInputBuf(inbuf)
		NotFound(outbuf)
		return
	}
	// Post exists
	var length int
	var buffered int
	// Read information from request head by lines
	buffered = inbuf.Buffered()
	for (buffered > 0) {
		line, _, _ := inbuf.ReadLine()
		str := string(line)
		println(str)
		// Check Content-Length
		if (strings.Contains(str, "Content-Length:")) {
			length, _ = strconv.Atoi(strings.Split(str, " ")[1])
			println("length :" + strconv.Itoa(length))
		}
		buffered = inbuf.Buffered()
		// Request head ended
		if (length > 0 && buffered == length+2) {
			break
		}
	}
	// Discard /r/n
	inbuf.Discard(2)
	// Get post data
	src := make([]byte, length)
	inbuf.Read(src)
	// Split body
	vec := strings.Split(string(src), "&")
	// Get name
	login := strings.TrimLeft(vec[0], "login=")
	// Get pass
	pass := strings.TrimLeft(vec[1], "pass=")
	// Check body
	if login == "3150102277" && pass == "102277" {
		writeOKHead(outbuf)
		var success []byte = []byte("<html><body>Login success!</body></html>")
		outbuf.Write([]byte("Content-Length: " + strconv.Itoa(len(success)) + "\n\n"))
		outbuf.Write(success)
	} else {
		writeOKHead(outbuf)
		var fail []byte = []byte("<html><body>Login failed</body></html>")
		outbuf.Write([]byte("Content-Length: " + strconv.Itoa(len(fail)) + "\n\n"))
		outbuf.Write(fail)
	}
}
func listen(c chan net.Conn, ln net.Listener) {
	for {
		// Wait for connection blocking
		conn, err := ln.Accept()
		if err != nil {
			// Handle error
			fmt.Println(err)
			continue
		}
		// Pass connection to main loop
		c <- conn
	}
}
func waitquit(c chan int) {
	var input string
	for {
		fmt.Scanln(&input)
		if input == "exit" {
			fmt.Println("exiting")
			c <- 1
		} else {
			fmt.Println("wrong input")
		}
	}

}
func writeOKHead(outbuf *bytes.Buffer) {
	outbuf.Write([]byte(head))
}
func NotFound(outbuf *bytes.Buffer) {
	outbuf.Write([]byte(notfound))
	var warning string = "<html><body>404 Not Found</body></html>"
	outbuf.Write([]byte("Content-Length: " + strconv.Itoa(len(warning)) + "\n\n"))
	outbuf.Write([]byte(warning))
}
