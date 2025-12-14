package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type PeerListItem struct {
	widget.Label
	OnTapped          func()
	OnTappedSecondary func()
}

func NewPeerListItem(text string, onTapped func(), onTappedSecondary func()) *PeerListItem {
	item := &PeerListItem{OnTapped: onTapped, OnTappedSecondary: onTappedSecondary}
	item.SetText(text)
	item.ExtendBaseWidget(item)
	return item
}

func (p *PeerListItem) Tapped(_ *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped()
	}
}

func (p *PeerListItem) TappedSecondary(_ *fyne.PointEvent) {
	if p.OnTappedSecondary != nil {
		p.OnTappedSecondary()
	}
}
