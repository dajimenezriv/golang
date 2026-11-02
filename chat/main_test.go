package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestNewChatServer(t *testing.T) {
	server := NewChatServer()
	if server == nil {
		t.Fatal("NewChatServer returned nil")
	}
}

func TestConnect(t *testing.T) {
	server := NewChatServer()

	// Test valid connection
	client, err := server.Connect("alice")
	if err != nil {
		t.Errorf("Failed to connect client: %v", err)
	}
	if client == nil {
		t.Error("Client is nil after successful connection")
	}

	// Test duplicate username
	_, err = server.Connect("alice")
	if err != ErrUsernameAlreadyTaken {
		t.Errorf("Expected ErrUsernameAlreadyTaken but got: %v", err)
	}

	// Test multiple valid connections
	for i := 0; i < 5; i++ {
		username := fmt.Sprintf("user%d", i)
		_, err := server.Connect(username)
		if err != nil {
			t.Errorf("Failed to connect user %s: %v", username, err)
		}
	}
}

func TestDisconnect(t *testing.T) {
	server := NewChatServer()

	client, _ := server.Connect("alice")
	server.Disconnect(client)

	// The username should be available again after disconnection
	newClient, err := server.Connect("alice")
	if err != nil {
		t.Errorf("Failed to connect with previously used username after disconnect: %v", err)
	}
	if newClient == nil {
		t.Error("Client is nil after successful reconnection")
	}
}

func TestBroadcast(t *testing.T) {
	server := NewChatServer()

	// Connect multiple clients
	clients := make([]*Client, 5)
	for i := 0; i < 5; i++ {
		username := fmt.Sprintf("user%d", i)
		client, _ := server.Connect(username)
		clients[i] = client

		// Start a goroutine to consume messages to prevent channel blocking
		go func(c *Client) {
			for {
				msg := c.Receive()
				if msg == "" {
					// Channel closed
					return
				}
			}
		}(client)
	}

	// Broadcast a message from one client
	testMessage := "Hello everyone!"
	server.Broadcast(clients[0], testMessage)

	// Give some time for messages to be delivered
	time.Sleep(100 * time.Millisecond)

	// Clean up
	for _, client := range clients {
		server.Disconnect(client)
	}
}

func TestPrivateMessage(t *testing.T) {
	server := NewChatServer()

	// Connect sender and recipient
	sender, _ := server.Connect("sender")
	recipient, _ := server.Connect("recipient")

	// Start a goroutine to receive messages for the recipient
	receivedMessages := make([]string, 0)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < 1; i++ { // Expect 1 message
			msg := recipient.Receive()
			if msg != "" {
				receivedMessages = append(receivedMessages, msg)
			}
		}
	}()

	// Send a private message
	testMessage := "This is a private message"
	err := server.PrivateMessage(sender, "recipient", testMessage)
	if err != nil {
		t.Errorf("Failed to send private message: %v", err)
	}

	// Wait for message to be received
	wg.Wait()

	// Verify the message was received
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(receivedMessages))
	} else if !strings.Contains(receivedMessages[0], testMessage) {
		t.Errorf("Message content mismatch. Expected to contain '%s', got '%s'", testMessage, receivedMessages[0])
	}

	// Test sending to non-existent recipient
	err = server.PrivateMessage(sender, "nonexistent", "Hello")
	if err != ErrRecipientNotFound {
		t.Errorf("Expected ErrRecipientNotFound but got: %v", err)
	}

	// Clean up
	server.Disconnect(sender)
	server.Disconnect(recipient)
}

func TestConcurrentOperations(t *testing.T) {
	server := NewChatServer()

	// Number of clients to create
	numClients := 20

	// Create clients
	clients := make([]*Client, numClients)
	for i := range numClients {
		username := fmt.Sprintf("user%d", i)
		client, _ := server.Connect(username)
		clients[i] = client

		// Start a goroutine to consume messages
		go func(c *Client) {
			for {
				msg := c.Receive()
				if msg == "" {
					// Channel closed or error
					return
				}
			}
		}(client)
	}

	// Perform concurrent operations
	var wg sync.WaitGroup
	wg.Add(numClients)

	for idx := range numClients {
		go func() {
			defer wg.Done()

			client := clients[idx]

			// Each client sends 5 broadcast messages
			for j := 0; j < 5; j++ {
				message := fmt.Sprintf("Broadcast %d from %d", j, idx)
				server.Broadcast(client, message)
			}

			// Each client sends 5 private messages to a random recipient
			for j := 0; j < 5; j++ {
				recipientIdx := rand.Intn(numClients)
				if recipientIdx == idx {
					// Avoid sending to self
					recipientIdx = (recipientIdx + 1) % numClients
				}

				recipientName := fmt.Sprintf("user%d", recipientIdx)
				message := fmt.Sprintf("Private %d from %d to %d", j, idx, recipientIdx)
				_ = server.PrivateMessage(client, recipientName, message)
			}
		}()
	}

	// Wait for all operations to complete
	wg.Wait()

	// Give some time for message delivery
	time.Sleep(200 * time.Millisecond)

	// Clean up
	for _, client := range clients {
		server.Disconnect(client)
	}
}

func TestDisconnectDuringOperation(t *testing.T) {
	server := NewChatServer()

	client1, _ := server.Connect("client1")
	client2, _ := server.Connect("client2")

	// Start receiving from client2
	done := make(chan bool)
	go func() {
		_ = client2.Receive() // This should not block forever if client1 is disconnected
		done <- true
	}()

	// Disconnect client1 during operation
	server.Disconnect(client1)

	// Try to send a message from client1 to client2
	err := server.PrivateMessage(client1, "client2", "Hello after disconnect")
	if err == nil {
		t.Error("Expected error when sending from disconnected client, but got nil")
	}

	// Clean up
	server.Disconnect(client2)

	// Ensure the receive goroutine completes
	select {
	case <-done:
		// Good, it completed
	case <-time.After(1 * time.Second):
		t.Error("Receive operation timed out after client disconnect")
	}
}
