package main

import (
	"fmt"
	"io"
	"bufio"
	"net"
	"strings"
	"sync"
	"time"
	"math/rand"
)

const MaxPlayerSlots = 8
const countDownTimer = 2
const lockInTimer = 3 // Players are locked in with their choices
var SelectedCourse = 0

type Client struct {
	Conn     net.Conn
	Username string
	Slot int
	IsPlayer bool // Observer if false
	Ready bool
	Course int
	Character int
	Loaded bool
}

type Lobby struct {
	Clients map[net.Conn]*Client
	VacantSlots []int // Slots that were occupied but are now vacant
	Mutex   sync.Mutex
	PlayerCount int
	UniqueCharacters bool
	StartGame bool
}

var GLobby = Lobby{
	Clients: make(map[net.Conn]*Client, 0),
	VacantSlots: make([]int, 0),
	PlayerCount: 0,
	UniqueCharacters: false,
	StartGame: false,
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

	if username == "" {
		fmt.Println("Invalid username")
		return
	}

	var client Client

	client.Conn = conn
	client.Username = username
	client.Ready = false
	
	// Max slots, assign observer role
	if GLobby.PlayerCount >= MaxPlayerSlots {
		client.IsPlayer = false
		message := fmt.Sprintf("%s joined as an observer! Hello!", username)
		GLobby.Clients[conn] = &client
		BroadcastStringTCP(message)
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
		GLobby.Clients[conn] = &client

		message := fmt.Sprintf("%s joined slot %d! Hello!", username, slot + 1)
		BroadcastStringTCP(message)
	}

}

func SelectCharacter(conn net.Conn, value int) {
	if GLobby.Clients[conn].IsPlayer {
		if GLobby.UniqueCharacters {
			for _, client := range GLobby.Clients {
				if client.Conn == conn {
					continue // Skip self
				}
				if client.Character == value {
					SendStringTCP(client, "This character has already been chosen")
					return
				}
			}
			GLobby.Clients[conn].Character = value
		} else { // Players can choose the same characters
			GLobby.Clients[conn].Character = value
		}
	}
}

func CourseVote(conn net.Conn, value int) {
	if GLobby.Clients[conn].IsPlayer {
		GLobby.Clients[conn].Course = value
	}
}

func ReadyUp(conn net.Conn, value bool) {
	if GLobby.Clients[conn].IsPlayer {
		if (value == true) {
			GLobby.Clients[conn].Ready = true
		} else {
			GLobby.Clients[conn].Ready = false
		}

		var count int = 0
		for _, client := range GLobby.Clients {
			if client.IsPlayer {
				if client.Ready {
					count++;
				}
			}
		}
		if count == (MaxPlayerSlots / 2) {
			StartCountdown()
		}
	}
}

func StartCountdown() {
	timer := time.NewTimer(countDownTimer * time.Second)
	<-timer.C

	timer2 := time.NewTimer(lockInTimer * time.Second)
	SelectedCourse = SelectCourse()
	SetPlayerType();
	<-timer2.C
	GLobby.StartGame = true
}

func SelectCourse() (int) {
	votes := make(map[int]int)
	// Select course
	for _, client := range GLobby.Clients {
		if client.IsPlayer {
			votes[client.Course]++
		}
	}

	// Find the course with the highest count
	var maxCount int
	var maxCourses []int
	for course, count := range votes {
		if count > maxCount {
			maxCount = count
			maxCourses = []int{course}
		} else if count == maxCount {
			maxCourses = append(maxCourses, course)
		}
	}

	// Randomly choose from tied courses if there are multiple with the same max count
	rand.Seed(time.Now().UnixNano())
	selectedCourse := maxCourses[rand.Intn(len(maxCourses))]
	fmt.Printf("Selected course: %d\n", selectedCourse)
	return selectedCourse
}

func Loaded(conn net.Conn) {
	GLobby.Clients[conn].Loaded = true

	for _, client := range GLobby.Clients {
		if (client.Loaded == false) {
			return
		}
	}
	// Start Game
	BroadcastPacket(LoadedPacket)
}

func LoadGame() {
	// Send packet to load the course
}

func Leave(conn net.Conn) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	client, ok := GLobby.Clients[conn]
	if !ok {
		return // Client not found
	}

	
	// Find and remove slot
	if client.IsPlayer {
		message := fmt.Sprintf("%s left, freeing slot %d\n", client.Username, client.Slot + 1)
		GLobby.VacantSlots = append(GLobby.VacantSlots, client.Slot)
		GLobby.PlayerCount--
		//fmt.Printf("%s left, freeing slot %d\n", client.Username, client.Slot + 1)
		BroadcastMessageTCP(GLobby.Clients[conn], message);
	} else {
		message := fmt.Sprintf("Observer %s left", client.Username)
		BroadcastMessageTCP(GLobby.Clients[conn], message);
	}
	delete(GLobby.Clients, conn)
}

func Message(conn net.Conn, message string) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	client, ok := GLobby.Clients[conn]
	if !ok {
		return // Client not found
	}
	msg := fmt.Sprintf("%s: %s\n", client.Username, message);
	BroadcastMessageTCP(GLobby.Clients[conn], msg);
}

func SetPlayerType() {

}

func BroadcastMessageTCP(caller *Client, msg string) {
	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		_, err := writer.WriteString(formatPacketTCP(MessagePacket, msg))
		if err != nil {
			fmt.Println("Error writing to client:", err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
		}
	}
	fmt.Println(msg)
}

func BroadcastStringTCP(message string) {
	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		
		_, err := writer.WriteString(formatPacketTCP(MessagePacket, message))
		if err != nil {
			fmt.Println("Error writing to client:", err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
		}
	}
	fmt.Printf("BroadcastString, %s\n", message)
}

func BroadcastPacket(packet int) {
	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		_, err := writer.WriteString(formatPacketTCP(packet, ""))
		if err != nil {
			fmt.Println("Error writing to client:", err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
		}
	}
}

func formatPacketTCP(packetType int, payload string) string {
    return fmt.Sprintf("%d:%s\n", packetType, payload)
}

func SendStringTCP(client *Client, message string) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	writer := bufio.NewWriter(client.Conn)
	_, err := writer.WriteString(message + "\n")
	if err != nil {
		fmt.Println("Error writing to client:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer:", err)
		return
	}

	fmt.Printf("Sent %s, to %s\n", message, client.Username)
}
