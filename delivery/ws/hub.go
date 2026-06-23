package ws

import (
	"sync"
)

// Frame data yang akan dikirim ke frontend
type BroadcastPayload struct {
	AuctionID int64   `json:"auction_id"`
	Amount    float64 `json:"amount"`
	User      string  `json:"user"` // nama pembeli terbaru (opsional)
	CreatedAt string  `json:"created_at"`
}

type Hub struct {
	// Map untuk mengelompokkan client berdasarkan AuctionID
	// key: auction_id, value: map berisi kumpulan pointer Client
	rooms      map[int64]map[*Client]bool
	broadcast  chan BroadcastPayload
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		broadcast:  make(chan BroadcastPayload),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.AuctionID] == nil {
				h.rooms[client.AuctionID] = make(map[*Client]bool)
			}
			h.rooms[client.AuctionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.rooms[client.AuctionID]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
					close(client.send)
					if len(connections) == 0 {
						delete(h.rooms, client.AuctionID)
					}
				}
			}
			h.mu.Unlock()

		case payload := <-h.broadcast:
			h.mu.RLock()
			connections := h.rooms[payload.AuctionID]
			for client := range connections {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					h.mu.Lock()
					delete(connections, client)
					h.mu.Unlock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Fungsi eksternal untuk dipanggil dari Use Case saat ada penawaran baru
func (h *Hub) BroadcastBid(payload BroadcastPayload) {
	h.broadcast <- payload
}
