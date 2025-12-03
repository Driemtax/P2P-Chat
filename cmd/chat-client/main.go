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
	peers := make([]string, 10)
	peers[0] = peer

	go Send(peer)
	go Reiceve(host)

	for {

	}

}

func printWelcome() {
	fmt.Println("######################################################################################")
	fmt.Println("                               Welcome to Chat 3000!                                  ")
	fmt.Println("######################################################################################")
	log.Println("Initialize Listener...")
}

func parseCMD() (string, string) {
	host := flag.String("port", "9000", "the port you want to listen on..")
	peer := flag.String("peer", "localhost:3000", "full remote address of the peer you want to connect to. <IP>:<PORT>")

	flag.Parse()
	port := ":" + *host

	return port, *peer
}

func Reiceve(host string) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	log.Println("Listening on Port", host)

	buf := make([]byte, maxBuffSize)
	for {
		// Read for new messages
		// This is a blocking function call, so it waits till a new message arrives at the port
		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s: %s", addr, string(buf[:size]))
	}

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
