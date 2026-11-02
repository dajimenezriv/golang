package main

import (
	"context"
	"errors"
	"sync"
)

type ConnectionStatus int

const (
	Connected ConnectionStatus = iota
	Disconnected
)

// Client represents a connected chat client
type Client struct {
	ctx      context.Context
	cancel   context.CancelFunc
	messages chan string

	mu       sync.Mutex
	username string
	status   ConnectionStatus
}

// Send sends a message to the client
func (c *Client) Send(message string) {
	select {
	case c.messages <- message:
	case <-c.ctx.Done():
	}
}

// Receive returns the next message for the client (blocking)
func (c *Client) Receive() string {
	select {
	case m := <-c.messages:
		return m
	case <-c.ctx.Done():
		return ""
	}
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

	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		ctx:      ctx,
		cancel:   cancel,
		messages: make(chan string),
		username: username,
		status:   Connected,
	}
	s.clients[username] = c
	return c, nil
}

// Disconnect removes a client from the chat server
func (s *ChatServer) Disconnect(client *Client) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.cancel()
	client.status = Disconnected
	delete(s.clients, client.username)
}

// Broadcast sends a message to all connected clients
func (s *ChatServer) Broadcast(sender *Client, message string) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.status == Disconnected {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		c.Send(message)
	}
}

// PrivateMessage sends a message to a specific client
func (s *ChatServer) PrivateMessage(sender *Client, recipient string, message string) error {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.status == Disconnected {
		return ErrClientDisconnected
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.clients[recipient]; ok {
		c.Send(message)
		return nil
	}
	return ErrRecipientNotFound
}

// Common errors that can be returned by the Chat Server
var (
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrRecipientNotFound    = errors.New("recipient not found")
	ErrClientDisconnected   = errors.New("client disconnected")
	// Add more error types as needed
)
