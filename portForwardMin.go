package main

import (
	"io"
	"log"
	"net"
	"os"
)

// https://gist.github.com/qhwa/cb9d3851450bff3b705e
func main() {
	go qhwa("127.0.0.1:2001", "192.168.100.10:80")
	go qhwa("127.0.0.1:2002", "192.168.100.10:80")
	go qhwa("127.0.0.1:2003", "192.168.100.10:80")
	go qhwa("127.0.0.1:2004", "192.168.100.10:80")
	<-make(chan bool)
}

func qhwa(s, d string) {
	ln, err := net.Listen("tcp", s)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error:", err)
			os.Exit(1)
		}
		go handleRequest(conn, d)
	}
}

// 处理请求
func handleRequest(conn net.Conn, d string) {
	proxy, err := net.Dial("tcp", d)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	go copyIO(conn, proxy)
	go copyIO(proxy, conn)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
