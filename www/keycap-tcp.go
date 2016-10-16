package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	b := make([]byte, 10)
	for true {
		n, err := conn.Read(b)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b[0:n]))
	}
}
