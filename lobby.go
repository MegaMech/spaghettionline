package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

const MaxPlayerSlots = 8

type Client struct {
	Conn     net.Conn
	Username string
	Slot int
	IsPlayer bool
}

type Lobby struct {
	Clients map[net.Conn]*Client
	VacantSlots []int // Slots that were occupied but are now vacant
	Mutex   sync.Mutex
	PlayerCount int
}

var GLobby = Lobby{
	Clients: make(map[net.Conn]*Client, 0),
	VacantSlots: make([]int, 0),
	PlayerCount: 0,
}

func CreateString(r io.Reader) (string, error) {
	var inbuf [64]byte
	var result strings.Builder

	for {
		n, err := r.Read(inbuf[:])
		if err != nil {
			fmt.Println("Failed to receive msg from the server!")
			break
		}
		result.Write(inbuf[:n])
		if inbuf[n] == '\000' {
			break
		}
	}
	return result.String(), nil
}

func Join(conn net.Conn, username string) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()
	
	if _, exists := GLobby.Clients[conn]; exists {
		fmt.Println("Client already exists")
		return
	}

	var client Client

	client.Conn = conn
	client.Username = username
	
	// Max slots, assign observer role
	if GLobby.PlayerCount >= MaxPlayerSlots {
		client.IsPlayer = false
		fmt.Printf("%s joined! Hi! Oh noes, max slots reached. You will be an observer %d\n", username, len(GLobby.Clients))
		// NotifyClients()
		return
	} else { // Assign player slot to the new client
		client.IsPlayer = true
		
		var slot int
		if len(GLobby.VacantSlots) > 0 {
			// Assign the first available slot
			slot = GLobby.VacantSlots[0]
			GLobby.VacantSlots = GLobby.VacantSlots[1:]
		} else {
			// Assign a new slot
			slot = GLobby.PlayerCount
		}
		GLobby.PlayerCount++

		client.Slot = slot

		fmt.Printf("%s joined! Hello! Slot %d\n", username, slot + 1)
	}

	GLobby.Clients[conn] = &client

	// Notify all clients about the new client
	NotifyClients()
}

func Leave(conn net.Conn) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	client, ok := GLobby.Clients[conn]
	if !ok {
		return // Client not found
	}

	delete(GLobby.Clients, conn)
	
	// Find and remove slot
	if client.IsPlayer {
		GLobby.VacantSlots = append(GLobby.VacantSlots, client.Slot)
		GLobby.PlayerCount--
		fmt.Printf("%s left, freeing slot %d\n", client.Username, client.Slot + 1)
	} else {
		fmt.Printf("Observer %s left\n", client.Username)
	}

	NotifyClients()
}

func Message(conn net.Conn, msg string) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	client, ok := GLobby.Clients[conn]
	if !ok {
		return // Client not found
	}
	fmt.Printf("%s says: %s\n", client.Username, msg)
	// Further processing can be done here, e.g., broadcasting the message
}

func NotifyClients() {
	//GLobby.Mutex.Lock()
	//defer GLobby.Mutex.Unlock()

	// var usernames []string
	// for _, client := range GLobby.Clients {
	// 	usernames = append(usernames, client.Username)
	// }

	// usernamesStr := strings.Join(usernames, ",")

	// for _, client := range GLobby.Clients {
	// 	writer := bufio.NewWriter(client.Conn)
	// 	_, err := writer.WriteString(usernamesStr + "\n")
	// 	if err != nil {
	// 		fmt.Println("Error notifying client:", err)
	// 		continue
	// 	}
	// 	writer.Flush()
	// }
	fmt.Println("End of Packet!")
}

