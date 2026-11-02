package main

import (
	"context"
	"errors"
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
	ctx      context.Context
	cancel   context.CancelFunc
	messages chan string
	username string
}

// Send sends a message to the client
func (c *Client) Send(message string) {
	select {
	case c.messages <- message:
	default: // buffer full: drop the message
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
		messages: make(chan string, 1024),
		username: username,
	}
	s.clients[username] = c
	return c, nil
}

// Disconnect removes a client from the chat server
func (s *ChatServer) Disconnect(client *Client) {
	client.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clients[client.username] == client {
		delete(s.clients, client.username)
	}
}

// Broadcast sends a message to all connected clients
func (s *ChatServer) Broadcast(sender *Client, message string) {
	if sender.ctx.Err() != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		if c != sender {
			c.Send(message)
		}
	}
}

// PrivateMessage sends a message to a specific client
func (s *ChatServer) PrivateMessage(sender *Client, recipient string, message string) error {
	if sender.ctx.Err() != nil {
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
