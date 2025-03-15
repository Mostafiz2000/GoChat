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

	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
)

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
	TimeStamp        string `json:"timestamp"`
}

type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceID   string `json:"deviceID"`
	IsLoggedIn bool   `json:"isLoggedIn"`
}

var users = map[string]string{}
var db *bolt.DB

func initDB() {
	var err error
	db, err = bolt.Open("localdb.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		return err
	})
}

func saveUser(user User) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		userData, _ := json.Marshal(user)
		return b.Put([]byte(user.Username), userData)
	})
}

func getUser(username string) (User, error) {
	var user User
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		userData := b.Get([]byte(username))
		if userData != nil {
			json.Unmarshal(userData, &user)
		}
		return nil
	})
	return user, nil
}
func getLoggedInUser() (User, bool) {
	var loggedInUser User
	var found bool
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		b.ForEach(func(k, v []byte) error {
			var user User
			json.Unmarshal(v, &user)
			if user.IsLoggedIn {
				loggedInUser = user
				found = true
				return nil
			}
			return nil
		})
		return nil
	})
	return loggedInUser, found
}

func loginUser(user User) bool {
	userData, _ := json.Marshal(user)
	resp, err := http.Post("http://localhost:8080/sign-in", "application/json", bytes.NewBuffer(userData))
	if err != nil {
		fmt.Println("⚠️ Login failed:", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func registerUserOnServer(user User) bool {
	userData, _ := json.Marshal(user)
	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(userData))
	if err != nil {
		fmt.Println("⚠️ Server registration failed:", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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

func main() {
	//initDB()
	reader := bufio.NewReader(os.Stdin)

	// if user, found := getLoggedInUser(); found {
	// 	fmt.Println("Welcome back,", user.Username)
	// 	connectWebSocket(user.Username)
	// 	return
	// }
	var username, password, deviceID string

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
			deviceID = getDeviceID()

			tempUser := User{
				Username: username,
				Password: password,
				DeviceID: deviceID,
			}

			if loginUser(tempUser) {
				fmt.Println("✅ Login successful!")
				//saveUser(User{Username: username, Password: password, IsLoggedIn: true})
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

			if _, exists := users[username]; exists {
				fmt.Println("⚠️ Username already taken! Choose another.")
				continue
			}

			fmt.Print("Create a password: ")
			password, _ = reader.ReadString('\n')
			password = strings.TrimSpace(password)

			newUser := User{
				Name:     name,
				Username: username,
				Password: password,
				DeviceID: deviceID,
			}

			// Save locally
			//saveUser(newUser)
			users[username] = password

			// Register on the server
			if registerUserOnServer(newUser) {
				fmt.Println("✅ Sign-up successful on server!")
			} else {
				fmt.Println("⚠️ Server registration failed, but local account created.")
			}
		}
	}
	connectWebSocket(username)
}

func connectWebSocket(username string) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Connection Error:", err)
	}
	defer conn.Close()

	user := User{Username: username}
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
			//saveUser(User{Username: username, IsLoggedIn: false})
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
