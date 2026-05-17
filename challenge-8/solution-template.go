// Package challenge8 contains the solution for Challenge 8: Chat Server with Channels.
package challenge8

import (
	"errors"
	"fmt"
	"sync"
	// Add any other necessary imports
)

// ---- エラー ----
var (
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrRecipientNotFound    = errors.New("recipient not found")
	ErrClientDisconnected   = errors.New("client disconnected")
)

// Client represents a connected chat client
type Client struct {
	// TODO: Implement this struct
	// Hint: username, message channel, mutex, disconnected flag
	Username     string
	incoming     chan string
	mu           sync.Mutex
	disconnected bool
}

func newClient(username string) *Client {
	return &Client{
		Username: username,
		incoming: make(chan string, 100),
	}
}

// Send sends a message to the client
// Send はスレッドセーフ、ノンブロッキング
func (c *Client) Send(message string) {
	// TODO: Implement this method
	// Hint: thread-safe, non-blocking send
	c.mu.Lock()
	if c.disconnected {
		c.mu.Unlock()
		return
	}
	ch := c.incoming
	c.mu.Unlock()

	// チャネルが閉じられても panic しない
	defer func() { recover() }()

	select {
	case ch <- message:
	default:
		// バッファが詰まったら別ゴルーチンで入れる（順序は保たれる）
		go func() {
			defer func() { recover() }()
			ch <- message
		}()
	}

}

// Receive returns the next message for the client (blocking)
// Receive はブロッキング読み出し
func (c *Client) Receive() string {
	// TODO: Implement this method
	// Hint: read from channel, handle closed channel
	msg, ok := <-c.incoming
	if !ok {
		return "" // チャネルが閉じられた場合は空文字を返す
	}
	return msg
}

// ChatServer manages client connections and message routing
type ChatServer struct {
	// TODO: Implement this struct
	// Hint: clients map, mutex
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewChatServer creates a new chat server instance
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]*Client),
	}
}

// Connect adds a new client to the chat server
func (s *ChatServer) Connect(username string) (*Client, error) {
	// TODO: Implement this method
	// Hint: check username, create client, add to map
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[username]; exists {
		return nil, ErrUsernameAlreadyTaken
	}
	c := newClient(username)
	s.clients[username] = c
	return c, nil
}

// Disconnect removes a client from the chat server
func (s *ChatServer) Disconnect(client *Client) {
	// TODO: Implement this method
	// Hint: remove from map, close channels
	if client == nil {
		return
	}
	s.mu.Lock()
	delete(s.clients, client.Username)
	s.mu.Unlock()

	client.mu.Lock()
	if !client.disconnected {
		client.disconnected = true
		close(client.incoming)
	}
	client.mu.Unlock()
}

// Broadcast sends a message to all connected clients
func (s *ChatServer) Broadcast(sender *Client, message string) {
	formatted := fmt.Sprintf("%s: %s", sender.Username, message)
	// TODO: Implement this method
	// Hint: format message, send to all clients
	s.mu.RLock()
	clients := make([]*Client, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	for _, c := range clients {
		// 自分にも送りたいなら条件を外す
		if c != sender {
			c.Send(formatted)
		}
	}
}

// PrivateMessage sends a message to a specific client
func (s *ChatServer) PrivateMessage(sender *Client, recipient string, message string) error {
	// TODO: Implement this method
	// Hint: find recipient, check errors, send message
	s.mu.RLock()
	target, ok := s.clients[recipient]
	s.mu.RUnlock()

	if !ok {
		return ErrRecipientNotFound
	}

	sender.mu.Lock()
	disconnected := sender.disconnected
	sender.mu.Unlock()
	if disconnected {
		return ErrClientDisconnected
	}

	formatted := fmt.Sprintf("%s->%s: %s", sender.Username, recipient, message)
	target.Send(formatted)

	return nil
}
