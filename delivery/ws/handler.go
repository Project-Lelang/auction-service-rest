package ws

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WsHandler struct {
	Hub *Hub
}

func NewWsHandler(hub *Hub) *WsHandler {
	return &WsHandler{Hub: hub}
}

func (h *WsHandler) ServeWs(c *gin.Context) {
	auctionIDStr := c.Param("auction_id")
	auctionID, err := strconv.ParseInt(auctionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Auction ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:       h.Hub,
		Conn:      conn,
		AuctionID: auctionID,
		send:      make(chan BroadcastPayload, 256),
	}
	client.Hub.register <- client

	// Jalankan rutin read & write secara asynchronous
	go client.WritePump()
	go client.ReadPump()
}
