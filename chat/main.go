package main

import (
	"errors"
	"fmt"
	"sync"
)

// Common errors that can be returned by the Chat Server
var (
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrRecipientNotFound    = errors.New("recipient not found")
	ErrClientDisconnected   = errors.New("client disconnected")
)

// Client represents a connected chat client
type Client struct {
	username string
	messages chan string

	done     chan struct{} // closed on disconnect
	doneOnce sync.Once
}

// Send sends a message to the client, dropping it if the buffer is full
func (c *Client) Send(message string) {
	if c.isDisconnected() {
		return
	}

	select {
	case c.messages <- message:
	default:
	}
}

// Receive returns the next message for the client (blocking).
// After a disconnect it drains queued messages, then returns "".
func (c *Client) Receive() string {
	select {
	case m := <-c.messages:
		return m
	case <-c.done:
	}

	select {
	case m := <-c.messages:
		return m
	default:
		return ""
	}
}

func (c *Client) isDisconnected() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *Client) disconnect() {
	c.doneOnce.Do(func() { close(c.done) })
}

// ChatServer manages client connections and message routing
type ChatServer struct {
	mu      sync.Mutex
	clients map[string]*Client
}

// NewChatServer creates a new chat server instance
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]*Client),
	}
}

// Connect adds a new client to the chat server
func (s *ChatServer) Connect(username string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[username]; ok {
		return nil, ErrUsernameAlreadyTaken
	}

	c := &Client{
		username: username,
		messages: make(chan string, 1024),
		done:     make(chan struct{}),
	}
	s.clients[username] = c
	return c, nil
}

// Disconnect removes a client from the chat server
func (s *ChatServer) Disconnect(client *Client) {
	client.disconnect()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clients[client.username] == client {
		delete(s.clients, client.username)
	}
}

// Broadcast sends a message to all connected clients, including the sender
func (s *ChatServer) Broadcast(sender *Client, message string) {
	if sender.isDisconnected() {
		return
	}

	formatted := fmt.Sprintf("%s: %s", sender.username, message)

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		c.Send(formatted)
	}
}

// PrivateMessage sends a message to a specific client. A nil error means the
// message was accepted, not delivered: it is dropped if the recipient's buffer
// is full, and may never be read.
func (s *ChatServer) PrivateMessage(sender *Client, recipient string, message string) error {
	if sender.isDisconnected() {
		return ErrClientDisconnected
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.clients[recipient]; ok {
		c.Send(fmt.Sprintf("%s: %s", sender.username, message))
		return nil
	}
	return ErrRecipientNotFound
}
