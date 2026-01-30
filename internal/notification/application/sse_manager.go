package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/domain"
	"log"
	"sync"
)

// SSEManager manages SSE client connections
type SSEManager struct {
	clients    map[int64][]port.SSEClientPort // userID -> clients (one user can have multiple connections)
	mu         sync.RWMutex
	register   chan port.SSEClientPort
	unregister chan port.SSEClientPort
	broadcast  chan *domain.SSEMessage
	stopCh     chan struct{}
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	manager := &SSEManager{
		clients:    make(map[int64][]port.SSEClientPort),
		register:   make(chan port.SSEClientPort, 100),
		unregister: make(chan port.SSEClientPort, 100),
		broadcast:  make(chan *domain.SSEMessage, 100),
		stopCh:     make(chan struct{}),
	}

	go manager.run()
	return manager
}

// run starts the main event loop
func (m *SSEManager) run() {
	for {
		select {
		case client := <-m.register:
			m.addClient(client)
		case client := <-m.unregister:
			m.removeClient(client)
		case message := <-m.broadcast:
			m.broadcastMessage(message)
		case <-m.stopCh:
			m.closeAllClients()
			return
		}
	}
}

// addClient adds a client to the manager
func (m *SSEManager) addClient(client port.SSEClientPort) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userID := client.GetUserID()
	m.clients[userID] = append(m.clients[userID], client)
	log.Printf("SSE: Client connected for user %d (total connections: %d)", userID, len(m.clients[userID]))
}

// removeClient removes a client from the manager
func (m *SSEManager) removeClient(client port.SSEClientPort) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userID := client.GetUserID()
	clients := m.clients[userID]
	for i, c := range clients {
		if c == client {
			m.clients[userID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	// Remove user entry if no more clients
	if len(m.clients[userID]) == 0 {
		delete(m.clients, userID)
	}

	client.Close()
	log.Printf("SSE: Client disconnected for user %d", userID)
}

// broadcastMessage sends a message to all connected clients
func (m *SSEManager) broadcastMessage(message *domain.SSEMessage) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, clients := range m.clients {
		for _, client := range clients {
			if !client.IsClosed() {
				if err := client.Send(message); err != nil {
					log.Printf("SSE: Failed to send broadcast message: %v", err)
				}
			}
		}
	}
}

// closeAllClients closes all client connections
func (m *SSEManager) closeAllClients() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, clients := range m.clients {
		for _, client := range clients {
			client.Close()
		}
	}
	m.clients = make(map[int64][]port.SSEClientPort)
}

// RegisterClient registers a new SSE client
func (m *SSEManager) RegisterClient(client port.SSEClientPort) {
	m.register <- client
}

// UnregisterClient removes an SSE client
func (m *SSEManager) UnregisterClient(client port.SSEClientPort) {
	m.unregister <- client
}

// SendToUser sends a notification to a specific user
func (m *SSEManager) SendToUser(userID int64, message *domain.SSEMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.clients[userID]
	if !exists || len(clients) == 0 {
		log.Printf("SSE: No active connections for user %d", userID)
		return nil // Not an error, user just not connected
	}

	for _, client := range clients {
		if !client.IsClosed() {
			if err := client.Send(message); err != nil {
				log.Printf("SSE: Failed to send message to user %d: %v", userID, err)
			}
		}
	}

	return nil
}

// Broadcast sends a notification to all connected clients
func (m *SSEManager) Broadcast(message *domain.SSEMessage) {
	m.broadcast <- message
}

// GetConnectedUsers returns the list of connected user IDs
func (m *SSEManager) GetConnectedUsers() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]int64, 0, len(m.clients))
	for userID := range m.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserConnected checks if a user is connected
func (m *SSEManager) IsUserConnected(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.clients[userID]
	return exists && len(clients) > 0
}

// Stop stops the SSE manager
func (m *SSEManager) Stop() {
	close(m.stopCh)
}
