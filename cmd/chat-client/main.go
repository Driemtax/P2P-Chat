package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const maxBuffSize = 128

func main() {
	printWelcome()
	host, peer := parseCMD()
	fmt.Println("Connecting to:", host, peer)

	go Send(peer)

	addr, err := net.ResolveUDPAddr("udp", host)
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
	log.Println("Initialize Listener...")
}

func parseCMD() (string, string) {
	host := flag.String("port", "8080", "the port you want to listen on..")
	peer := flag.String("peer", "localhost:3000", "full remote address of the peer you want to connect to. <IP>:<PORT>")

	flag.Parse()
	port := ":" + *host

	return port, *peer
}

func Send(peer string) {
	addr, err := net.ResolveUDPAddr("udp", peer)
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

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Fatal(err)
		}
	}
}
