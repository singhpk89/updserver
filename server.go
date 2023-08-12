/*
Usage:

	go run server.go

See https://gist.github.com/winlinvip/e8665ba888e2fd489ccd5a449fadfa73
See https://stackoverflow.com/a/70576851/17679565
See https://github.com/ossrs/srs/issues/2843
*/
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	serverPort := 5001
	if len(os.Args) > 1 {
		if v, err := strconv.Atoi(os.Args[1]); err != nil {
			fmt.Printf("Invalid port %v, err %v", os.Args[1], err)
			os.Exit(-1)
		} else {
			serverPort = v
		}
	}

	addr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	server, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Listen err %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("Listen at %v\n", addr.String())

	for {
		p := make([]byte, 1024)
		nn, raddr, err := server.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Read err  %v", err)
			continue
		}

		msg := p[:nn]
		fmt.Printf("Received %v %s\n", raddr, msg)

		go func(conn *net.UDPConn, raddr *net.UDPAddr, msg []byte) {
			_, err := conn.WriteToUDP([]byte(fmt.Sprintf("Pong: %s", msg)), raddr)
			if err != nil {
				fmt.Printf("Response err %v", err)
			}
		}(server, raddr, msg)
	}
}
