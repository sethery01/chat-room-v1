/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 * 	October 24, 2025
 *	Information used in project from: https://pkg.go.dev/
 *	client/main.go
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

// Sends a message to the server and waits for a response
// message is content to send
// conn is current connection to server
// returns the message from the server as a Go string
func sendAndReceive(conn net.Conn, message []byte) string {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return "0"
	}

	// Listen for the server response
	buffer := make([]byte, SIZE_OF_BUFF)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return "0"
	}

	// Parse response
	response := buffer[0:bytesRead]
	data := string(response)

	return data
}

// Attempts to log the user in
// conn is current server connection
// command is the raw input from the command line to send to the server
// returns true on success and false on failure
func login(conn net.Conn, command string) bool {
	// // Send the login message
	message := []byte(command)
	data := sendAndReceive(conn, message)

	// Validate login
	if data != "1" {
		fmt.Println("> Denied. Username or password incorrect.")
		return false
	}
	fmt.Println("> You are logged in!")
	return true
}

// logout closes connection conn and returns true on success
func logout(conn net.Conn) bool {
	// Send the logout message
	message := []byte("logout")
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// newuser will try and register a newuser on the server
// conn is the current connection to the server
// command is the raw from the command line to send to the server
// returns true upson success and false on failure
func newuser(conn net.Conn, command string) bool {
	// Send the newuser message
	message := []byte(command)
	data := sendAndReceive(conn, message)

	// Validate registration
	if data != "1" {
		fmt.Println("> Denied. User account already exists.")
		return false
	}
	fmt.Println("> New account created! Please login.")
	return true
}

// send sends a message to the server and prints the response
// conn is the current connection to the server
// command is the raw input from the command line to send to the server
func send(conn net.Conn, command string) {
	// Send the message to be echoed to the user
	message := []byte(command)
	data := sendAndReceive(conn, message)
	fmt.Println("> " + data)
}

// start is the main loop of the client program
// start listens for commands until the user logs out or hits "control + c"
func start(conn net.Conn) {
	fmt.Println("******************************************************************")
	fmt.Print("Hello! Welcome to Seth Ek's chatbot V1.\n\nAvailable commands:\n")
	fmt.Print("login \"UserID\" \"Password\"\nnewuser \"UserID\" \"Password\"\nsend \"message\"\nlogout\n")
	fmt.Print("\nPlease enter commands as shown above. You must begin with login.\n")
	fmt.Println("******************************************************************")

	// I/O reader
	reader := bufio.NewReader(os.Stdin)
	loggedIn := false

	// Wait for input from user
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')     // Get raw input
		inputString := strings.TrimSpace(input) // Trim whitespace here
		command := strings.Fields(inputString)  // strings.Fields creates a slice of strings seperated by a space

		// no input, don't panic
		if len(command) == 0 {
			continue
		}

		// Execute the command
		switch command[0] {
		
		// case login uses the login() func to execute this command
		case "login":
			if loggedIn {
				fmt.Println("> You are already logged in.")
			} else if len(command) == 3 {
				loggedIn = login(conn, inputString)
			} else {
				fmt.Println("> You must provided a username and password.")
			}

		// case newuser uses the newuser() func to execute this command
		case "newuser":
			if loggedIn {
				fmt.Println("> Denied. You cannot create a new user while logged in.")
			} else if len(command) != 3 {
				fmt.Println("> You must provided a username and password.")
			} else if len(command[1]) < 3 || len(command[1]) > 32 {
				fmt.Println("> Your username must be between 3 and 32 characters.")
			} else if len(command[2]) < 4 || len(command[2]) > 8 {
				fmt.Println("> Your password must be between 4 and 8 characters.")
			} else {
				newuser(conn, inputString)
			}
		
		// case send uses the send() func to execute this command
		case "send":
			if !loggedIn {
				fmt.Println("> Denied. Please login before sending a message.")
			} else if len(command) < 2 {
				fmt.Println("> You must include a message.")
			} else if len(strings.Join(command[2:], " ")) > 256 {
				fmt.Println("> Your message must be 1-256 characters long.")
			} else {
				send(conn, inputString)
			}
		
		// case logout uses the logout() func to execute this command
		case "logout":
			if loggedIn {
				loggedOut := logout(conn)
				if loggedOut {
					fmt.Println("> See you next time!")
					return
				}
				fmt.Println("> Error logging out.")
			} else {
				fmt.Println("> You must login before logging out.")
			}
		default:
			fmt.Println("> Invalid command")
		}
	}
}

func main() {
	// Connect to the socket via tcp
	conn, err := net.Dial("tcp", SOCKET)
	if err != nil {
		fmt.Println("Please start the server first.")
		log.Fatal(err) // This will kill thr program
	}

	// Run the app
	start(conn)

	conn.Close()
}
