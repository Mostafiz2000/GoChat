package main

import (
	"bufio"

	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func main() {
	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Connection Error:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	// Ask for username
	fmt.Print("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Send username to server
	user := User{Username: username}
	err = conn.WriteJSON(user)
	if err != nil {
		log.Fatal("Error sending username:", err)
	}

	// Start listening for incoming messages in a separate goroutine
	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}
			fmt.Printf("\nNew message from %s: %s\n", msg.SenderUsername, msg.Content)
			fmt.Print("Press 'M' to send a message, or 'Q' to quit: ") // Reprint prompt after incoming message
		}
	}()

	// Main loop for sending messages
	for {
		fmt.Print("Press 'M' to send a message, or 'Q' to quit: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "Q" || input == "q" {
			fmt.Println("Goodbye!")
			break
		} else if input == "M" || input == "m" {
			// Get recipient username
			fmt.Print("Enter the receiver's username: ")
			receiver, _ := reader.ReadString('\n')
			receiver = strings.TrimSpace(receiver)

			// Get message content
			fmt.Print("Enter your message: ")
			messageContent, _ := reader.ReadString('\n')
			messageContent = strings.TrimSpace(messageContent)

			// Create and send message
			message := Message{
				SenderUsername:   username,
				ReceiverUsername: receiver,
				Content:          messageContent,
			}

			err := conn.WriteJSON(message)
			if err != nil {
				log.Println("Error sending message:", err)
			} else {
				fmt.Println("Message sent!")
			}
		}
	}
}
