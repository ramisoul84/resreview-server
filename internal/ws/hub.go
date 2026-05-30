package ws

import (
	"encoding/json"
	"sync"
)

type Client struct {
	Hub    *Hub
	UserID string
	Name   string
	Color  string
	Send   chan []byte
}

type userInfo struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}

type presenceMsg struct {
	Type   string     `json:"type"`
	Online int        `json:"online"`
	Users  []userInfo `json:"users"`
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	online     map[string]*userInfo
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		online:     make(map[string]*userInfo),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.online[client.UserID] = &userInfo{
				UserID: client.UserID,
				Name:   client.Name,
				Color:  client.Color,
			}
			h.mu.Unlock()
			h.broadcastPresence()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.online[client.UserID] = nil
				delete(h.online, client.UserID)
			}
			h.mu.Unlock()
			h.broadcastPresence()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					delete(h.clients, client)
					close(client.Send)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) buildPresenceMsg() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]userInfo, 0, len(h.online))
	for _, info := range h.online {
		if info != nil {
			users = append(users, *info)
		}
	}
	return json.Marshal(presenceMsg{
		Type:   "presence",
		Online: len(users),
		Users:  users,
	})
}

func (h *Hub) Broadcast(data []byte) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	h.broadcast <- dataCopy
}

func (h *Hub) broadcastPresence() {
	msg, err := h.buildPresenceMsg()
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			delete(h.clients, client)
			close(client.Send)
		}
	}
}
