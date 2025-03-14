package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

var DB *sql.DB

// Initialize database connection and run migrations
func InitDB() {
	var err error
	DB, err = sql.Open("mysql", "root:@tcp(localhost:3306)/chatdb")
	if err != nil {
		log.Fatal("Error connecting to MySQL:", err)
	}

	// Check database connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✅ Database connected successfully!")

	// Run database migrations
	migrateDB()
}

func migrateDB() {
	// Create users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			deviceID VARCHAR(255) DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal("Error migrating users table:", err)
	}

	// Create messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			sender_id INT NOT NULL,
			receiver_id INT NOT NULL,
			content TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		log.Fatal("Error migrating messages table:", err)
	}

	fmt.Println("✅ Database migration completed!")
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"deviceID"`
}

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
	TimeStamp        string `json:"timestamp"`
}

func RegisterUser(db *sql.DB, name, username, password string) (int, error) {
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&existingID)
	if err == nil {
		return 0, errors.New("username already exists")
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := db.Exec("INSERT INTO users (name, username, password) VALUES (?, ?, ?)", name, username, password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func AuthenticateUser(db *sql.DB, username, password string) (int, error) {
	var userID int
	var storedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &storedPassword)
	if err != nil {
		return 0, errors.New("invalid username or password")
	}

	if storedPassword != password {
		return 0, errors.New("invalid username or password")
	}

	return userID, nil
}

func SaveMessage(db *sql.DB, senderID, receiverID int, content string) error {
	_, err := db.Exec("INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)", senderID, receiverID, content)
	return err
}

func GetUserIDByUsername(db *sql.DB, username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
