package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *ChatUI) makeSidebar() fyne.CanvasObject {
	ui.PeerList = widget.NewList(
		func() int {
			ui.peersMutex.RLock()
			defer ui.peersMutex.RUnlock()
			return len(ui.Peers)
		},
		func() fyne.CanvasObject {
			// Why?
			return NewPeerListItem("Template", nil, nil)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			ui.peersMutex.RLock()
			// If peer list shrinked, id could be out of bounds
			if id >= len(ui.Peers) {
				ui.peersMutex.RUnlock()
				return
			}
			peer := ui.Peers[id]
			ui.peersMutex.RUnlock()
			pItem := item.(*PeerListItem)
			pItem.SetText(peer.Alias)

			// Left Click: Open Direct Message Chat
			pItem.OnTapped = func() {
				// TODO: Open Direct Chat window
			}

			// Right Click: Context menue
			pItem.OnTappedSecondary = func() {
				menu := fyne.NewMenu("Optionen",
					fyne.NewMenuItem("Chat öffnen", func() {
						// TODO: Open direct chat
					}),
					fyne.NewMenuItem("Peer Info", func() {
						ui.ShowPeerInfoDialog(peer)
					}),
				)

				// Show menue at mouse click position
				widget.ShowPopUpMenuAtPosition(menu, ui.Window.Canvas(),
					fyne.CurrentApp().Driver().AbsolutePositionForObject(pItem))
			}
		},
	)

	// Scan Button & Progress Bar
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// --- SCAN BUTTON ---
	var scanBtn *widget.Button
	onSend := func() {
		// UI logic for network scan
		scanBtn.Hide()
		progressBar.SetValue(0)
		progressBar.Show()

		// Handle progress bar in goroutine
		anim := fyne.NewAnimation(5*time.Second, func(v float32) {
			progressBar.SetValue(float64(v))

			if v >= 1.0 {
				progressBar.Hide()
				scanBtn.Show()
			}
		})
		anim.Curve = fyne.AnimationLinear
		anim.Start()

		// Trigger real broadcast in background
		go ui.PeerManager.ScanNetwork()
	}

	scanBtn = widget.NewButton("Netzwerk yeeten", onSend)

	// --- ADD BUTTON ---
	addPeerBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
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

	})

	return container.NewBorder(addPeerBtn, container.NewStack(progressBar, scanBtn), nil, nil, ui.PeerList)
}

func (ui *ChatUI) ShowPeerInfoDialog(peer PeerData) {
	aliasEntry := widget.NewEntry()
	aliasEntry.SetText(peer.Alias)

	dialog.ShowForm("Peer bearbeiten", "Speichern", "Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("IP Adresse", widget.NewLabel(peer.IP)),
			widget.NewFormItem("Nickname", aliasEntry),
		},
		func(submitted bool) {
			if submitted {
				ui.PeerManager.SetAlias(peer.ID, aliasEntry.Text)
				ui.PeerList.Refresh()
			}
		}, ui.Window)
}
