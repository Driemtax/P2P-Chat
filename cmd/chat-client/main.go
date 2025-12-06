package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"chat-client/internal"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const maxBuffSize = 128

var globalPort string = "9000"

func main() {
	printWelcome()
	globalPort, peer := parseCMD()
	peers := &internal.Peers{}
	peers.Add(peer)

	// We open a global port for sending and receiving messages. This is a threadsafe operation.
	addr, err := net.ResolveUDPAddr("udp", globalPort)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

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
	go func(peers *internal.Peers, conn *net.UDPConn) {
		defer conn.Close()
		for msg := range sendChan {
			// directly send the message as a byte array via udp
			err = peers.Broadcast(conn, msg)
			if err != nil {
				timestamp := time.Now().Format("2006/01/02 15:04:05")
				chatLog, _ := chatHistory.Get()
				errMsg := fmt.Sprintf("\n%s [System]: Konnte Nachricht nicht senden (Empfänger offline?): %v", timestamp, err)
				newLog := chatLog + errMsg
				chatHistory.Set(newLog)
				log.Printf("\033[31m[Fehler]: Konnte Nachricht nicht senden (Empfänger offline?): %v\033[0m", err)

			}
		}
	}(peers, conn)

	// GO routine for reiceving a message
	go func(conn *net.UDPConn) {
		chatLog, _ := chatHistory.Get()
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		newLog := chatLog + timestamp + " System: Now Listening on Port " + globalPort + "\n"
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
	}(conn)

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
