package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"chat-client/internal"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const maxBuffSize = 128

var globalCache []byte = make([]byte, maxBuffSize)

func main() {
	printWelcome()
	host, peer := parseCMD()
	fmt.Println("Connecting to:", host, peer)
	peers := []net.UDPAddr{}
	internal.AddPeer(peers, peer)

	a := app.New()
	window := a.NewWindow("Chat 3000")

	// Chat history widget
	chatHistory := binding.NewString()
	chatHistory.Set("Willkommen im Chat\n")

	history := widget.NewLabelWithData(chatHistory)

	// Input text field
	input := widget.NewEntry()
	input.PlaceHolder = "Nachricht tippen..."

	// Create a channel for communcation between send routine and ui thread
	sendChan := make(chan string, 10)

	onSend := func() {
		msg := input.Text
		if msg != "" {
			sendChan <- msg
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			chatLog, _ := chatHistory.Get()
			chatHistory.Set(chatLog + "\n" + timestamp + " You: " + msg + "\n")
			input.SetText("")
		}
	}

	input.OnSubmitted = func(s string) {
		onSend()
	}

	// Send-Button
	send := widget.NewButton("Senden", onSend)

	content := container.NewVBox(
		history,
		input,
		send,
	)

	// GO routine for sending a message
	go func(peer string) {
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
		for msg := range sendChan {
			// directly send the message as a byte array via udp
			_, err = conn.Write([]byte(msg))
			if err != nil {
				timestamp := time.Now().Format("2006/01/02 15:04:05")
				chatLog, _ := chatHistory.Get()
				errMsg := fmt.Sprintf("\n%s [System]: Konnte Nachricht nicht senden (Empfänger offline?): %v", timestamp, err)
				newLog := chatLog + errMsg
				chatHistory.Set(newLog)
				log.Printf("\033[31m[Fehler]: Konnte Nachricht nicht senden (Empfänger offline?): %v\033[0m", err)

			}
		}
	}(peer)

	// GO routine for reiceving a message
	go func(host string) {
		addr, err := net.ResolveUDPAddr("udp", host)
		if err != nil {
			log.Fatal(err)
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()
		chatLog, _ := chatHistory.Get()
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		newLog := chatLog + timestamp + " System: Now Listening on Port " + host + "\n"
		chatHistory.Set(newLog)

		buf := make([]byte, maxBuffSize)
		for {
			// Read for new messages
			// This is a blocking function call, so it waits till a new message arrives at the port
			size, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			// Print it to the chat widget
			chatLog, _ = chatHistory.Get()
			timestamp = time.Now().Format("2006/01/02 15:04:05")
			msg := fmt.Sprintf("\n%s %s: %s\n", timestamp, addr.String(), string(buf[:size]))
			newLog := chatLog + msg
			chatHistory.Set(newLog)
		}
	}(host)

	window.SetContent(content)
	window.ShowAndRun()
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

		// Print it to the chat widget
		log.Printf("%s: %s\n", addr, string(buf[:size]))
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
