// Copyright (C) Philip Schlump, 2016-2017.

package WebSocketChat

//NOT OK

// hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients    map[*Client]bool // Registered clients.
	broadcast  chan []byte      // Inbound messages from the clients.
	register   chan *Client     // Register requests from the clients.
	unregister chan *Client     // Unregister requests from clients.
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
