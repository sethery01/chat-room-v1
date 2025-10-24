/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 * 	October 24, 2025
 *	Information used in project from: https://pkg.go.dev/
 *	server/main.go
***********************************************************************/
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	SOCKET       = "127.0.0.1:10740"
	SIZE_OF_BUFF = 1024
)

// Sends a message to the connected client
// message is the content to send
// conn is the current connection
func sendMessage(conn net.Conn, message []byte) {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
	}
}

// Evaluates whether the user is registered and checks password
// username is the UserID of the account
// password is the Password
// newuser is a flag determining whether to attempt login or newuser validation
// returns true if a user exists already OR  true if username and password correct
// returns false if user does not exist OR false if username and password incorrect
func validateUser(username, password string, newuser bool) bool {
	// Open the users file
	file, err := os.Open("users.txt")
	if err != nil {
		log.Println(err)
		return false
	}
	defer file.Close() // Close file after func exit

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Parse the txt file for the username and password
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // skip empty lines
		}
		line = strings.TrimPrefix(line, "(")
		line = strings.TrimSuffix(line, ")")
		user := strings.Split(line, ",")
		if len(user) < 2 {
			continue // skip bad lines
		}
		user[0] = strings.TrimSpace(user[0])
		user[1] = strings.TrimSpace(user[1])

		// Check if username and password match for login case
		if user[0] == username && user[1] == password && !newuser {
			return true
		}

		// Check if user exists for newuser case
		if user[0] == username && newuser {
			return true
		}
	}
	return false
}

// Helper function to complete login flow
// command is the parsed input from the client-side
func login(command []string) bool {
	username := command[1]
	password := command[2]
	return validateUser(username, password, false)
}

// Determines whether the requested newuser can be added
// command is the parsed input from the client-side
// returns true on successful newuser, false otherwise
func newuser(command []string) bool {
	// set username/password for readability
	username := command[1]
	password := command[2]

	// Check if the user exists... if not continue registration
	userExists := validateUser(username, password, true)
	if userExists {
		log.Printf("User %s already exists.\n", username)
		return false
	} else {
		// Open the users file for appending
		file, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return false
		}
		defer file.Close() // Close file after func exit

		// If file is empty, don't add a newline
		fileInfo, _ := file.Stat()
		prefix := ""
		if fileInfo.Size() > 0 {
			prefix = "\n"
		}
		
		// Write the new user to the EOF
		user := fmt.Sprintf("%s(%s, %s)", prefix, username, password)
		_, err = file.Write([]byte(user))
		if err != nil {
			log.Println(err)
			return false
		}
	}

	// Success
	log.Print("User added: " + username)
	return true
}

// Function to handle incoming connections
// conn accepted in main
func handleConnection(conn net.Conn) {
	log.Println("New connection from: " + conn.RemoteAddr().String())
	defer conn.Close() // Close connection upon function exit
	loggedIn := false  // Initally not logged in
	activeUser := ""   // No user yet... random conn

	// Listen until connection terminated
	for {
		// Read in data sent by the connection
		buffer := make([]byte, SIZE_OF_BUFF)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}

		// Parse request
		request := buffer[0:bytesRead]
		data := string(request)
		command := strings.Fields(data)

		// Execute the command sent from the client
		switch command[0] {

		// login case sends 1 on success, 0 on failure
		case "login":
			loggedIn = login(command)
			var message []byte
			if loggedIn {
				activeUser = command[1]
				log.Println("User logged in as:", activeUser)
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn, message)

		// newuser sends 1 on success, 0 on failure
		case "newuser":
			created := newuser(command)
			var message []byte
			if created {
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn, message)

		// broadcast message to server output and client if logged in
		case "send":
			if loggedIn {
				data = data[5:]
				message := []byte(fmt.Sprintf("%s: %s", activeUser, data))
				log.Println(string(message))
				sendMessage(conn, message)
			} else {
				log.Println("User is not logged in.")
			}

		// Should never error here, but only logout if logged in
		// Terminates connection upon logout and this thread closes
		case "logout":
			if loggedIn {
				log.Println(activeUser, "logout.")
				log.Println("Terminating connection: " + conn.RemoteAddr().String())
				return
			} else {
				log.Println("User chose logout but nobody is logged in.")
			}
		default:
			log.Println("Invalid command")
		}
	}
}

func main() {
	// Ensure users.txt exists before server starts
	if _, err := os.Stat("users.txt"); os.IsNotExist(err) {
		file, err := os.Create("users.txt")
		if err != nil {
			log.Fatalf("Failed to create users.txt: %v", err)
		}
		file.Close()
		log.Println("Created users.txt")
	}

	// Start the server
	ln, err := net.Listen("tcp", SOCKET)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ln.Close() // Close upon exiting main
	log.Println("Server is listening on " + ln.Addr().String())

	// Listen for and handle connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
