# P2P-Chat
A small Peer-to-Peer Chat program using UDP

## TO-DO:
- [x] Multicasting implementieren
- [x] Peers manuell hinzufügen
- [ ] Gossip protokoll, um neue Peers zu teilen
- [x] UI überlegen, wie man allen oder nur ausgewählten Peers Nachrichten sendet
- [x] MaxBuffer auf 1024 und erstes Byte für Message Type
- [x] Broadcast Button, um alle im Netzwerk zu pingen und somit neue Clients zu finden
  1. HSplit anstelle von BorderLayout als Top-Level-Container
  2. PeerManager erweitern, damit er UI über neue Peers benachrichtigen kann
  3. Button & Progress Bar zum Scan via Broadcast
- [x] Ping und Pong immer abchecken, ob man den Peer schon kennt -> Brauche ich hier eigene Goroutine?
- [ ] UI für Chatrooms -> seitliche Leiste mit Chatinterfaces und Bindings pro Peer um Verlauf zu speichern
- [ ] Dockerize für lokales Testen

## Testing
For testing you can add other localhost IP-Adresses since localhost is 127.0.0.0/8. On Linux this can be done with the following command:
```bash
sudo ip addr add 127.0.0.2/8 dev lo
```
