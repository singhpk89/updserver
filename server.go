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
	"time"

	csgolog "github.com/janstuemmel/csgo-log"
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
		fmt.Printf("Received %v : %s\n", raddr, msg)

		go response(server, raddr, msg)
	}
}

func response(udpServer net.PacketConn, addr net.Addr, buf []byte) {
	line := string(buf)
	data := parse(line)
	data = data + "\n"
	writeToFile(data)
	udpServer.WriteTo(cleanByteData([]byte(data)), addr)
}

func parse(line string) string {
	var msg csgolog.Message
	msg, err := csgolog.Parse(line)
	if err != nil {
		fmt.Println(err)
		return "Invalid data"
	} else {
		// get json non-htmlescaped
		jsn := csgolog.ToJSON(msg)
		fmt.Println(jsn)
		return jsn
	}
}

func cleanByteData(input []byte) []byte {
	cleanedData := make([]byte, 0, len(input))
	for _, b := range input {
		if b != 0 {
			cleanedData = append(cleanedData, b)
		}
	}
	return cleanedData
}

func writeToFile(content string) {
	currentDate := time.Now()
	dateString := currentDate.Format("2006-01-02") // Format: YYYY-MM-DD
	// currentDate := time.Now().Format("2023-08-01") // Format: YYYY-MM-DD

	// Create the file name with today's date
	fileName := dateString + "-log.log"
	// Open the file in append mode
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Append the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error appending content:", err)
		return
	}
}
