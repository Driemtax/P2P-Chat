package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const maxBuffSize = 128

func main() {
	printWelcome()

	addr, err := net.ResolveUDPAddr("udp", ":2345")
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
		log.Printf("Received %s from %s \n", string(buf[:size]), addr)
		_, err = ln.WriteToUDP([]byte(time.Now().String()), addr)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func printWelcome() {
	fmt.Println("######################################################################################")
	fmt.Println("                               Welcome to Chat 3000!                                  ")
	fmt.Println("######################################################################################")
	log.Println("Starting Listener on Port 2345...")
}
