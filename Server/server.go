package main

import (
	"database/sql"
	"encoding/json"
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
	http.HandleFunc("/sign-in", signInHandler)
	http.HandleFunc("/register", registerUserHandler)

	fmt.Println("🚀 Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func signInHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := models.AuthenticateUser(user.Username, user.Password)
	if err != nil {
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "User login successfully!",
		"userID":  userID,
	}
	json.NewEncoder(w).Encode(response)
}

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := models.RegisterUser(user.Name, user.Username, user.Password, user.DeviceID)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "User registered successfully!",
		"userID":  userID,
	}
	json.NewEncoder(w).Encode(response)
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

	var userID int

	existingUser, err := models.GetUserByUsername(user.Username)
	if err != nil {
		// User not found, proceed with registration
		log.Printf("User %s not found, registering...", existingUser.Name)
		_, err = models.RegisterUser(user.Name, user.Username, user.Password, user.DeviceID)
		if err != nil {
			log.Println("Error registering user:", err)
			conn.WriteJSON(map[string]string{"error": "Registration failed"})
			return
		}
		conn.WriteJSON(map[string]string{"success": "User registered successfully!"})
	} else {
		// User found, proceed with authentication
		userID, err = models.AuthenticateUser(user.Username, user.Password)
		if err != nil {
			log.Println("Authentication failed:", err)
			conn.WriteJSON(map[string]string{"error": "Invalid credentials"})
			return
		} else {
			user.ID = userID
			mutex.Lock()
			clients[user.Username] = conn
			mutex.Unlock()
		}

	}

	log.Printf("User %s logged in with ID %d\n", user.Username, userID)

	// Listen for incoming messages
	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Error reading message:", err)
			break
		}

		receiverID, err := models.GetUserIDByUsername(msg.ReceiverUsername)
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
