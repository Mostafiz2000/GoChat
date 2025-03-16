package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/gorilla/websocket"
)

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
	TimeStamp        string `json:"timestamp"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"deviceID"`
}

var serverIP string
var serverPort string
var tempUser User

func main() {
	// Ensure the correct number of arguments are provided
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run client.go <server_ip> <server_port>")
		return
	}

	serverIP = os.Args[1]
	serverPort = os.Args[2]
	serverURL := fmt.Sprintf("http://%s:%s", serverIP, serverPort)
	wsURL := fmt.Sprintf("ws://%s:%s/ws", serverIP, serverPort)

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
			bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
			password = string(bytePassword)
			fmt.Println("")
			deviceID := getDeviceID()

			tempUser = User{
				Username: username,
				Password: password,
				DeviceID: deviceID,
			}

			if loginUser(tempUser, serverURL) {
				fmt.Println("✅ Login successful!")
				break
			} else {
				fmt.Println("❌ Invalid username or password. Try again.")
			}
		} else if choice == "2" {
			fmt.Print("Enter your name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Choose a username: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)

			fmt.Print("Create a password: ")
			bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
			password = string(bytePassword)
			fmt.Println("")

			deviceID := getDeviceID()

			newUser := User{
				Username: username,
				Password: password,
				DeviceID: deviceID,
			}

			if registerUserOnServer(newUser, serverURL) {
				fmt.Println("✅ Sign-up successful!")
			} else {
				fmt.Println("⚠️ Server registration failed.")
			}
		}
	}

	connectWebSocket(tempUser, wsURL)
}

func loginUser(user User, serverURL string) bool {
	userData, _ := json.Marshal(user)
	resp, err := http.Post(fmt.Sprintf("%s/sign-in", serverURL), "application/json", bytes.NewBuffer(userData))
	if err != nil {
		fmt.Println("⚠️ Login failed:", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func registerUserOnServer(user User, serverURL string) bool {
	userData, _ := json.Marshal(user)
	resp, err := http.Post(fmt.Sprintf("%s/register", serverURL), "application/json", bytes.NewBuffer(userData))
	if err != nil {
		fmt.Println("⚠️ Registration failed:", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func connectWebSocket(user User, wsURL string) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Connection Error:", err)
	}
	defer conn.Close()

	err = conn.WriteJSON(user)
	if err != nil {
		log.Fatal("Error sending username:", err)
	}

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

	reader := bufio.NewReader(os.Stdin)
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
				SenderUsername:   user.Username,
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

func getDeviceID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.HardwareAddr != nil {
			return iface.HardwareAddr.String() // MAC address
		}
	}
	return "unknown"
}
