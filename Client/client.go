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
	TimeStamp        string `json:"timestamp"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"deviceID"`
}

var users = map[string]string{} // Simulating a simple user database (username -> password)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var username, password string

	for {
		fmt.Println("1. Login")
		fmt.Println("2. Sign Up")
		fmt.Print("Choose an option (1 or 2): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "1" {
			fmt.Print("Enter your username: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)

			fmt.Print("Enter your password: ")
			password, _ = reader.ReadString('\n')
			password = strings.TrimSpace(password)

			if storedPassword, exists := users[username]; exists && storedPassword == password {
				fmt.Println("Login successful!")
				break
			} else {
				fmt.Println("Invalid username or password. Try again.")
			}

		} else if choice == "2" {
			fmt.Print("Enter your name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Choose a username: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)

			if _, exists := users[username]; exists {
				fmt.Println("Username already taken! Choose another.")
				continue
			}

			fmt.Print("Create a password: ")
			password, _ = reader.ReadString('\n')
			password = strings.TrimSpace(password)

			users[username] = password
			fmt.Println("Sign-up successful! You can now log in.")
		}
	}

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Connection Error:", err)
	}
	defer conn.Close()

	// Send user info to server
	user := User{Username: username}
	err = conn.WriteJSON(user)
	if err != nil {
		log.Fatal("Error sending username:", err)
	}

	// Listen for incoming messages
	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}
			fmt.Printf("\n%s: %s\n", msg.SenderUsername, msg.Content)
			fmt.Print("Press 'M' to send a message, or 'Q' to quit: ")
		}
	}()

	for {
		fmt.Print("Press 'M' to send a message, or 'Q' to quit: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "Q" || input == "q" {
			fmt.Println("Goodbye!")
			break
		} else if input == "M" || input == "m" {
			fmt.Print("Enter the receiver's username: ")
			receiver, _ := reader.ReadString('\n')
			receiver = strings.TrimSpace(receiver)

			fmt.Print("Enter your message: ")
			messageContent, _ := reader.ReadString('\n')
			messageContent = strings.TrimSpace(messageContent)

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
