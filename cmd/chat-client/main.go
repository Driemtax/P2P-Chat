package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const maxBuffSize = 128
const peerAddress = "localhost:10234"
const port = ":2345"

func main() {
	printWelcome()

	go Send()

	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Listening on ", ln.LocalAddr().String())

	buf := make([]byte, maxBuffSize)
	for {
		// First read from the port we listen for new messages
		size, addr, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s: %s \n", addr, string(buf[:size]))
	}

}

func printWelcome() {
	fmt.Println("######################################################################################")
	fmt.Println("                               Welcome to Chat 3000!                                  ")
	fmt.Println("######################################################################################")
	log.Println("Starting Listener on Port 10234...")
}

func Send() {
	addr, err := net.ResolveUDPAddr("udp", peerAddress)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	for {
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		log.Println("You:", input)

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Fatal(err)
		}
	}
}
