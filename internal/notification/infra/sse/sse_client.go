package sse

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/domain"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// SSEClient represents an SSE client connection
type SSEClient struct {
	userID int64
	writer io.Writer
	closed bool
	mu     sync.RWMutex
}

// NewSSEClient creates a new SSE client
func NewSSEClient(userID int64, writer io.Writer) *SSEClient {
	return &SSEClient{
		userID: userID,
		writer: writer,
		closed: false,
	}
}

// Send sends a message to the client
func (c *SSEClient) Send(message *domain.SSEMessage) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// SSE format: "data: <json>\n\n"
	_, err = fmt.Fprintf(c.writer, "event: notification\ndata: %s\n\n", string(data))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// Flush if possible
	if flusher, ok := c.writer.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	return nil
}

// SendHeartbeat sends a heartbeat to keep the connection alive
func (c *SSEClient) SendHeartbeat() error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	_, err := fmt.Fprintf(c.writer, ": heartbeat\n\n")
	if err != nil {
		return fmt.Errorf("failed to write heartbeat: %w", err)
	}

	if flusher, ok := c.writer.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	return nil
}

// Close closes the client connection
func (c *SSEClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

// IsClosed checks if the client connection is closed
func (c *SSEClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// GetUserID returns the user ID of the client
func (c *SSEClient) GetUserID() int64 {
	return c.userID
}
