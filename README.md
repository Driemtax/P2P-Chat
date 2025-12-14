# P2P-Chat
A small Peer-to-Peer Chat program using UDP

## TO-DO:
- [x] Multicasting implementieren
- [x] Peers manuell hinzufügen
- [ ] Gossip protokoll, um neue Peers zu teilen
- [x] UI überlegen, wie man allen oder nur ausgewählten Peers Nachrichten sendet
- [ ] UI für Chatrooms -> seitliche Leiste mit Chatinterfaces und Bindings pro Peer um Verlauf zu speichern
- [x] MaxBuffer auf 1024 und erstes Byte für Message Type
- [ ] Broadcast Button, um alle im Netzwerk zu pingen und somit neue Clients zu finden
  1. HSplit anstelle von BorderLayout als Top-Level-Container
  2. PeerManager erweitern, damit er UI über neue Peers benachrichtigen kann
  3. Button & Progress Bar zum Scan via Broadcast
- [x] Ping und Pong immer abchecken, ob man den Peer schon kennt -> Brauche ich hier eigene Goroutine?
