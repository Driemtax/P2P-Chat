package ui

import (
	"chat-client/internal"
	"chat-client/internal/config"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UI model
type PeerData struct {
	IP    string
	Alias string
}

type ChatUI struct {
	App         fyne.App
	Window      fyne.Window
	PeerManager *internal.PeerManager

	Peers             []PeerData // copy of peers to not have to use mutex. Is this a good idea??
	snackbarContainer *fyne.Container
	ChatHistory       binding.String

	// channels
	SendChan chan string
}

func NewChatUI(pm *internal.PeerManager) *ChatUI {
	a := app.New()
	w := a.NewWindow(config.Title)

	ui := &ChatUI{
		App:         a,
		Window:      w,
		PeerManager: pm,
		Peers:       []PeerData{},
		ChatHistory: binding.NewString(),
		SendChan:    make(chan string, 10), // buffered for 10 messages
	}

	ui.ChatHistory.Set("Willkommen im Chat\n")
	ui.setupContent()
	return ui
}

func (ui *ChatUI) Start() {
	ui.Window.ShowAndRun()
}

func (ui *ChatUI) setupContent() {
	// Create sidebar
	// TODO

	// Chat Area
	history := widget.NewLabelWithData(ui.ChatHistory)
	history.Wrapping = fyne.TextWrapWord
	scrollContainer := container.NewScroll(history)

	// Set new Listener to chat binding so that the scroll Container always scrolls to the bottom when chatHistory changes
	ui.ChatHistory.AddListener(binding.NewDataListener(func() {
		scrollContainer.ScrollToBottom()
	}))

	// Input text field
	input := widget.NewEntry()
	input.PlaceHolder = "Nachricht tippen..."

	// Sending messages
	onSend := func() {
		msg := input.Text
		if msg != "" {
			ui.SendChan <- msg
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			chatLog, _ := ui.ChatHistory.Get()
			ui.ChatHistory.Set(chatLog + "\n" + timestamp + " You: " + msg + "\n")
			input.SetText("")
			ui.Window.Canvas().Focus(input)
		}
	}

	input.OnSubmitted = func(s string) {
		onSend()
	}
	sendBtn := widget.NewButton("Senden", onSend)
	inputArea := container.NewVBox(input, sendBtn)

	// TODO Sidebar here
	// sidebar := widget.NewLabel("Sidebar Placeholder")
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
						err := ui.PeerManager.Add(peerEntry.Text, aliasEntry.Text)

						if err != nil {
							dialog.ShowError(err, ui.Window)
						} else {
							// Log success to chat
							timestamp := time.Now().Format("2006/01/02 15:04:05")
							chatLog, _ := ui.ChatHistory.Get()
							ui.ChatHistory.Set(chatLog + "\n" + timestamp + " [System]: Peer " + peerEntry.Text + " hinzugefügt.\n")
						}
					}
				}, ui.Window)
		}),
	)

	// Layout of all components
	// HSplit for Sidebar | Chat
	chatContainer := container.NewBorder(nil, inputArea, nil, nil, scrollContainer)

	split := container.NewHSplit(toolbar, chatContainer)
	split.SetOffset(0.3) // 30% sidebar, adjust maybe

	ui.Window.SetContent(split)
	ui.Window.Resize(fyne.NewSize(800, 600))
	ui.Window.CenterOnScreen()

	// focus input on startup
	ui.Window.Canvas().Focus(input)

	// focus input on reopen chat window
	ui.App.Lifecycle().SetOnEnteredForeground(func() {
		ui.Window.Canvas().Focus(input)
	})

}
