package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var db *sql.DB

// Store active WebSocket connections
var clients = make(map[string]*websocket.Conn)
var mutex = sync.Mutex{} // To prevent race conditions

func main() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/chatdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	var user User
	if err := conn.ReadJSON(&user); err != nil {
		log.Println("Error reading user data:", err)
		return
	}

	// Register or get existing user
	userID, err := registerUser(user.Username)
	if err != nil {
		log.Println("Error registering user:", err)
		return
	}

	// Send userID back to client after successful registration
	user.ID = userID
	if err := conn.WriteJSON(user); err != nil {
		log.Println("Error sending user ID back:", err)
		return
	}

	log.Printf("User %s connected with ID %d\n", user.Username, userID)

	// Store WebSocket connection in clients map
	mutex.Lock()
	clients[user.Username] = conn
	mutex.Unlock()

	// Listen for incoming messages
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Get receiver ID
		receiverID, err := getUserIDByUsername(msg.ReceiverUsername)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "No user with this username exists"})
			continue
		}

		// Save message to the database
		if err := saveMessage(userID, receiverID, msg.Content); err != nil {
			log.Println("Error saving message:", err)
		}

		// Check if the receiver is online
		mutex.Lock()
		receiverConn, online := clients[msg.ReceiverUsername]
		mutex.Unlock()

		if online {
			// Forward message to online recipient
			err := receiverConn.WriteJSON(msg)
			if err != nil {
				log.Println("Error forwarding message:", err)
			}
		}

		// Send confirmation back to sender
		conn.WriteJSON(map[string]string{"status": "Message sent!"})
	}

	// Remove user from active clients on disconnect
	mutex.Lock()
	delete(clients, user.Username)
	mutex.Unlock()
}

func registerUser(username string) (int, error) {
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&existingID)
	if err == nil {
		return existingID, nil
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := db.Exec("INSERT INTO users (username) VALUES (?)", username)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func getUserIDByUsername(username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func saveMessage(senderID, receiverID int, content string) error {
	_, err := db.Exec("INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)", senderID, receiverID, content)
	return err
}
