package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"

	models "GoChat/Model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var db *sql.DB
var clients = make(map[string]*websocket.Conn)
var mutex = sync.Mutex{}

func main() {
	models.InitDB()

	http.HandleFunc("/ws", handleConnections)

	fmt.Println("🚀 Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	var user models.User
	if err := conn.ReadJSON(&user); err != nil {
		log.Println("Error reading user data:", err)
		return
	}

	// Authenticate user
	userID, err := models.AuthenticateUser(db, user.Username, user.Password)
	if err != nil {
		log.Println("Authentication failed:", err)
		conn.WriteJSON(map[string]string{"error": "Invalid credentials"})
		return
	}

	// Store user session
	user.ID = userID
	mutex.Lock()
	clients[user.Username] = conn
	mutex.Unlock()

	log.Printf("User %s logged in with ID %d\n", user.Username, userID)

	// Listen for incoming messages
	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Error reading message:", err)
			break
		}

		receiverID, err := models.GetUserIDByUsername(db, msg.ReceiverUsername)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "User not found"})
			continue
		}

		// Save message to the database
		if err := models.SaveMessage(db, userID, receiverID, msg.Content); err != nil {
			log.Println("Error saving message:", err)
		}

		// Forward message if recipient is online
		mutex.Lock()
		receiverConn, online := clients[msg.ReceiverUsername]
		mutex.Unlock()

		if online {
			err := receiverConn.WriteJSON(msg)
			if err != nil {
				log.Println("Error forwarding message:", err)
			}
		}

		// Confirm message sent
		conn.WriteJSON(map[string]string{"status": "Message sent!"})
	}

	// Remove user on disconnect
	mutex.Lock()
	delete(clients, user.Username)
	mutex.Unlock()
}
