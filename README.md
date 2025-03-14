# GoChat
## Overview

This project is a scalable Golang CLI application designed with a robust server-client architecture. It enables seamless peer-to-peer communication between clients, ensuring efficient and direct messaging without unnecessary intermediaries. The system is built for performance, scalability, and reliability, making it well-suited for real-time chat applications.

## Prerequisites

Before starting, ensure you have the following installed:

- [Visual Studio Code](https://code.visualstudio.com/download) – Recommended code editor for development.  
- [Golang](https://go.dev/doc/install) – The Go programming language (ensure it's the latest stable version).  
- [XAMPP](https://www.apachefriends.org/download.html) – Includes MySQL for database management.  
- **WebSockets** – For real-time bidirectional communication.  
- **gqlgen** – For type-safe client-server interactions.
  
**Verify Go installation** by running:  

   ```bash
   go version
   ```
## 1. Clone the Repository

Clone the source code to your local machine using the following command:
```bash
git clone <repository-url>
```
## 2. Project Setup
### 1. Setup Visual studio Code
Once you have cloned the repository, navigate to the project folder and open the GoChat folder in Visual Studio Code.

### 2. **Create an SQL Script for Database Setup (`chatdb`)**  
create a `chatdb` database on mysql database.

## 3. Build the Project
To build the project, follow these steps:  

```bash
# Navigate to the project directory
cd GoChat

# Install dependencies
go mod tidy

# Build the application
go build -o gochat
```

## 4. Run the Project  

To run the project, you need to start the server first and then run the client. Follow these steps:  
### 1. Run XAMPP  

Before running the GoChat server, ensure that **XAMPP** is running to handle MySQL.  

1. Open the **XAMPP Control Panel**.  
2. Start the **Apache** and **MySQL** services. 

### 2. Run the Server  

In the project directory, run the server using the following command:  

```bash
# Start the server
go run server.go
```
### 3. Run the Client  

 ```bash
# Start the server
go run client.go
```


