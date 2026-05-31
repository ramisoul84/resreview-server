package ws

import (
	"encoding/json"
	"sync"
)

type Client struct {
	Hub       *Hub
	UserID    string
	Name      string
	Color     string
	VersionID string
	Send      chan []byte
	closed    bool
}

type userInfo struct {
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	VersionID string `json:"versionId"`
}

type presenceMsg struct {
	Type   string     `json:"type"`
	Online int        `json:"online"`
	Users  []userInfo `json:"users"`
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	online     map[string]int
	userInfo   map[string]*userInfo
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	stop       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		online:     make(map[string]int),
		userInfo:   make(map[string]*userInfo),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		stop:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.online[client.UserID]++
			if h.online[client.UserID] == 1 {
				h.userInfo[client.UserID] = &userInfo{
					UserID:    client.UserID,
					Name:      client.Name,
					Color:     client.Color,
					VersionID: client.VersionID,
				}
			}
			h.mu.Unlock()
			h.broadcastPresence()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if !client.closed {
					client.closed = true
					close(client.Send)
				}
				h.online[client.UserID]--
				if h.online[client.UserID] <= 0 {
					delete(h.online, client.UserID)
					delete(h.userInfo, client.UserID)
				}
			}
			h.mu.Unlock()
			h.broadcastPresence()

		case msg := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					delete(h.clients, client)
					if !client.closed {
						client.closed = true
						close(client.Send)
					}
					h.online[client.UserID]--
					if h.online[client.UserID] <= 0 {
						delete(h.online, client.UserID)
						delete(h.userInfo, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case <-h.stop:
			h.mu.Lock()
			for client := range h.clients {
				if !client.closed {
					client.closed = true
					close(client.Send)
				}
			}
			h.clients = make(map[*Client]bool)
			h.online = make(map[string]int)
			h.userInfo = make(map[string]*userInfo)
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.stop)
}

func (h *Hub) buildPresenceMsg() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]userInfo, 0, len(h.userInfo))
	for _, info := range h.userInfo {
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

func (h *Hub) SetClientVersion(c *Client, versionID string) {
	h.mu.Lock()
	c.VersionID = versionID
	if info, ok := h.userInfo[c.UserID]; ok {
		info.VersionID = versionID
	}
	h.mu.Unlock()
	h.broadcastPresence()
}

func (h *Hub) Broadcast(data []byte) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	select {
	case h.broadcast <- dataCopy:
	default:
	}
}

func (h *Hub) broadcastPresence() {
	msg, err := h.buildPresenceMsg()
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			delete(h.clients, client)
			if !client.closed {
				client.closed = true
				close(client.Send)
			}
			h.online[client.UserID]--
			if h.online[client.UserID] <= 0 {
				delete(h.online, client.UserID)
				delete(h.userInfo, client.UserID)
			}
		}
	}
}
