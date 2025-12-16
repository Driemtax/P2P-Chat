package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"chat-client/internal"
	"chat-client/internal/ui"
)

// Protocol config
const maxBuffSize = 1024

var globalPort string = "9000"

// Window config
const width = 800
const height = 600

func main() {
	printWelcome()
	globalPort, _ := parseCMD()
	peerManager := internal.NewPeerManager(globalPort)

	chatUi := ui.NewChatUI(peerManager)

	peerManager.OnPeerListUpdated = func(peers []internal.Peer) {
		chatUi.UpdatePeerList(peers)
	}
	peerManager.StartHeartbeatLoop()

	// GO routine for sending a message
	go func(chatUI *ui.ChatUI) {
		defer peerManager.CloseSocket()
		for msg := range chatUI.SendChan {
			// add the header to the message
			payload := internal.PrepareMessage(internal.MsgTypeBroadcast, msg)
			// directly send the message as a byte array via udp
			err := peerManager.Broadcast(payload)
			if err != nil {
				timestamp := time.Now().Format("2006/01/02 15:04:05")
				chatLog, _ := chatUI.ChatHistory.Get()
				errMsg := fmt.Sprintf("\n%s [System]: Konnte Nachricht nicht senden (Empfänger offline?): %v", timestamp, err)
				newLog := chatLog + errMsg
				chatUI.ChatHistory.Set(newLog)
				log.Printf("\033[31m[Fehler]: Konnte Nachricht nicht senden (Empfänger offline?): %v\033[0m", err)
			}
		}
	}(chatUi)

	// GO routine for reiceving a message
	go func(chatUI *ui.ChatUI) {
		chatLog, _ := chatUI.ChatHistory.Get()
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		newLog := chatLog + timestamp + " System: Now Listening on Port " + globalPort + "\n"
		chatUI.ChatHistory.Set(newLog)

		buf := make([]byte, maxBuffSize)
		for {
			// Read for new messages
			// This is a blocking function call, so it waits till a new message arrives at the port
			size, addr, err := peerManager.ReadConn(buf)
			if err != nil {
				log.Fatal(err)
			}

			msgType, message, err := internal.HandleIncomingPacket(peerManager, addr, buf[:size])

			if msgType == internal.MsgTypeBroadcast || msgType == internal.MsgTypeUnicast {
				// Print it to the chat widget
				chatLog, _ = chatUI.ChatHistory.Get()
				timestamp = time.Now().Format("2006/01/02 15:04:05")
				alias := peerManager.GetAlias(addr)
				msg := fmt.Sprintf("\n%s %s: %s\n", timestamp, alias, string(message))
				newLog := chatLog + msg
				chatUI.ChatHistory.Set(newLog)
			}
		}
	}(chatUi)

	// Run the app and start the infinite loop of chatting with your peerd -> you can nerver escape (except you just close the app)
	chatUi.Start()
}

func printWelcome() {
	fmt.Println("######################################################################################")
	fmt.Println("                               Welcome to Chat 3000!                                  ")
	fmt.Println("######################################################################################")
	log.Println("Initialize Listener...")
}

func parseCMD() (string, string) {
	port := flag.String("port", "9000", "the port you want to listen on..")
	peer := flag.String("peer", "localhost:3000", "full remote address of the peer you want to connect to. <IP>:<PORT>")
	bind := flag.String("bind", "", "IP-Adrress you want to bind to (for local testing)")

	flag.Parse()
	bindAddress := ":" + *port

	if *bind != "" {
		bindAddress = *bind + ":" + *port
	}

	return bindAddress, *peer
}
