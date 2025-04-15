package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type MetricsMessage struct {
	Type        string    `json:"type"`
	Dimension   string    `json:"dimension"`
	Value       string    `json:"value"`
	SuccessRate float64   `json:"success_rate"`
	Timestamp   time.Time `json:"timestamp"`
}

type AlertMessage struct {
	Type            string    `json:"type"`
	ID              string    `json:"id"`
	Dimension       string    `json:"dimension"`
	Value           string    `json:"value"`
	CurrentRate     float64   `json:"current_rate"`
	PreviousRate    float64   `json:"previous_rate"`
	DropPercentage  float64   `json:"drop_percentage"`
	Timestamp       time.Time `json:"timestamp"`
	RootCause       string    `json:"root_cause,omitempty"`
	Confidence      float64   `json:"confidence,omitempty"`
	Recommendations []string  `json:"recommendations,omitempty"`
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastMetrics(metrics *MetricsMessage) {
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics: %v", err)
		return
	}
	h.Broadcast <- data
}

func (h *Hub) BroadcastAlert(alert *AlertMessage) {
	data, err := json.Marshal(alert)
	if err != nil {
		log.Printf("Error marshaling alert: %v", err)
		return
	}
	h.Broadcast <- data
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
