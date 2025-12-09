package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"chat-client/internal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	peerManager := internal.NewPeerManager()

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
	history.Wrapping = fyne.TextWrapWord
	scrollContainer := container.NewScroll(history)

	// Set new Listener to chat binding so that the scroll Container always scrolls to the bottom when chatHistory changes
	chatHistory.AddListener(binding.NewDataListener(func() {
		scrollContainer.ScrollToBottom()
	}))

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
			window.Canvas().Focus(input)
		}
	}

	input.OnSubmitted = func(s string) {
		onSend()
	}

	// Send-Button
	send := widget.NewButton("Senden", onSend)

	// Toolbar for adding peers
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			// Create input field for the dialog
			peerEntry := widget.NewEntry()
			peerEntry.PlaceHolder = "127.0.0.1:3000"
			aliasEntry := widget.NewEntry()
			aliasEntry.PlaceHolder = "Gib einen Nickname ein.."

			// SHow a form dialog
			dialog.ShowForm("Neuen Peer hinzufügen", "Hinzufügen", "Abbrechen",
				[]*widget.FormItem{
					widget.NewFormItem("Adresse", peerEntry),
					widget.NewFormItem("Nickname", aliasEntry),
				},
				func(submitted bool) {
					if submitted && peerEntry.Text != "" {
						err := peerManager.Add(peerEntry.Text, aliasEntry.Text)

						if err != nil {
							dialog.ShowError(err, window)
						} else {
							// Log success to chat
							timestamp := time.Now().Format("2006/01/02 15:04:05")
							chatLog, _ := chatHistory.Get()
							chatHistory.Set(chatLog + "\n" + timestamp + " [System]: Peer " + peerEntry.Text + " hinzugefügt.\n")
						}
					}
				}, window)
		}),
	)

	// Wrap Input field and button in a container
	inputArea := container.NewVBox(
		input,
		send,
	)

	content := container.NewBorder(
		toolbar,         // Top
		inputArea,       // Bottom
		nil,             // Left
		nil,             // Right
		scrollContainer, // Center
	)

	// GO routine for sending a message
	go func(peers *internal.PeerManager, conn *net.UDPConn) {
		defer conn.Close()
		for msg := range sendChan {
			// add the header to the message
			payload := internal.PrepareMessage(internal.MsgTypeBroadcast, msg)
			// directly send the message as a byte array via udp
			err = peerManager.Broadcast(conn, payload)
			if err != nil {
				timestamp := time.Now().Format("2006/01/02 15:04:05")
				chatLog, _ := chatHistory.Get()
				errMsg := fmt.Sprintf("\n%s [System]: Konnte Nachricht nicht senden (Empfänger offline?): %v", timestamp, err)
				newLog := chatLog + errMsg
				chatHistory.Set(newLog)
				log.Printf("\033[31m[Fehler]: Konnte Nachricht nicht senden (Empfänger offline?): %v\033[0m", err)
			}
		}
	}(peerManager, conn)

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

			msgType, message, err := internal.HandleIncomingPacket(conn, peerManager, addr, buf[:size])

			if msgType == internal.MsgTypeBroadcast || msgType == internal.MsgTypeUnicast {
				// Print it to the chat widget
				chatLog, _ = chatHistory.Get()
				timestamp = time.Now().Format("2006/01/02 15:04:05")
				alias := peerManager.GetAlias(addr)
				msg := fmt.Sprintf("\n%s %s: %s\n", timestamp, alias, string(message))
				newLog := chatLog + msg
				chatHistory.Set(newLog)
			}
		}
	}(conn)

	window.SetContent(content)
	window.Resize(fyne.NewSize(width, height))
	window.CenterOnScreen()

	// Focus the input field on startup and when app is reopend after tabbing out
	window.Canvas().Focus(input)
	a.Lifecycle().SetOnEnteredForeground(func() {
		window.Canvas().Focus(input)
	})

	// Run the app and start the infinite loop of chatting with your peerd -> you can nerver escape (except you just close the app)
	window.ShowAndRun()
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
