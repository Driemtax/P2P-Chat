package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const maxBuffSize = 128

var globalCache []byte = make([]byte, maxBuffSize)

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

// Reiceves a message via udp. Listens on the port specified in host. Printet message is capped to maxBuffSize
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

		// 1. tempo save the current input buffer
		//_, err = os.Stdin.ReadAt(globalCache, 0)
		//if err != nil {
		//	log.Fatal(err)
		//}
		// 2. Delete input
		//fmt.Printf("\r\033[K") // Deletes current row and jumps to next one
		// 3. print reiceved message
		log.Printf("%s: %s", addr, string(buf[:size]))
		// 4. reprint input buffer to continue typing
		//_, err = os.Stdin.WriteAt(globalCache, 0)
		//if err != nil {
		//	log.Fatal(err)
		//}
	}

}

// Send sends the Stdin Buffer to the configured udp address (<IP>:<PORT>)
func Send(peer string) {
	addr, err := net.ResolveUDPAddr("udp", peer)
	if err != nil {
		log.Fatal(err)
	}

	// Open a socket for sending (random empty socket is chosen cause of laddr=nil)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	for {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		fmt.Printf("%s You: ", timestamp)
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Printf("\033[31m[Fehler]: Konnte Nachricht nicht senden (Empfänger offline?): %v\033[0m", err)
		}
	}
}
