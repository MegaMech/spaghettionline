package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"math/rand"
)

const MaxPlayerSlots = 8
const countDownTimer = 2
const lockInTimer = 3 // Players are locked in with their choices
var LockedIn = false
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

type NetworkClient struct { // For sending to the clients
	Username string
	Slot int
	IsPlayer bool
	IsAI bool
	Character int
	/* Does the client own this client?
	 * This is important so that the client knows which player it is controlling
	 * The client should always be using gPlayerOne. However, this is needed to know where to place everyone
	 * on the starting line.
	 */
	HasAuthority bool
}

type Lobby struct {
	Clients map[net.Conn]*Client
	VacantSlots []int // Slots that were occupied but are now vacant
	Mutex   sync.Mutex
	PlayerCount int
	UniqueCharacters bool
	StartSession bool
}

var GLobby = Lobby{
	Clients: make(map[net.Conn]*Client, 0),
	VacantSlots: make([]int, 0),
	PlayerCount: 0,
	UniqueCharacters: false,
	StartSession: false,
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

	// Too late, the game is already starting
	if (LockedIn) {
		return
	}
	
	if _, exists := GLobby.Clients[conn]; exists {
		fmt.Println("Client already exists")
		return
	}

	if username == "" || !isValidUsername(username) {
		fmt.Println("Invalid username: ", username)
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

func SetCharacter(conn net.Conn, value []byte) {
	if GLobby.Clients[conn].IsPlayer {
		character := int(value[0])

		if GLobby.UniqueCharacters {
			for _, client := range GLobby.Clients {
				if client.Conn == conn {
					continue // Skip self
				}
				if client.Character == character {
					SendMessageToPlayer(client, "This character has already been chosen")
					return
				}
			}
			GLobby.Clients[conn].Character = character
		} else { // Players can choose the same characters
			GLobby.Clients[conn].Character = character
			SendMessageToPlayer(GLobby.Clients[conn], "Chosen character: "+string(character));
		}
	}
}

func CourseVote(conn net.Conn, value int) {
	if GLobby.Clients[conn].IsPlayer {
		GLobby.Clients[conn].Course = value
	}
}

func ReadyUp(conn net.Conn, value []byte) {
	if GLobby.Clients[conn].IsPlayer {

		if value[0] == 1 {
			GLobby.Clients[conn].Ready = true
		} else {
			GLobby.Clients[conn].Ready = false
		}

		var count int = 0
		var currentPlayers int = 0
		for _, client := range GLobby.Clients {
			if client.IsPlayer {
				if client.Ready {
					count++;
				}
				currentPlayers++
			}
		}
		//if count == (currentPlayers / 2) {
		if count > 0 { // <-- Debug. Real ^
			StartCountdown()
		}
	}
}

func StartCountdown() {
	timer := time.NewTimer(countDownTimer * time.Second)
	fmt.Printf("Starting countdown %ds\n", countDownTimer);
	<-timer.C

	LockedIn = true
	fmt.Printf("Final countdown %ds\n", lockInTimer);
	
	timer2 := time.NewTimer(lockInTimer * time.Second)
	SelectCourse()
	BroadcastPlayerSlots();
	<-timer2.C
	GLobby.StartSession = true
	BroadcastPacket(StartSessionPacket)
}

func SelectCourse() {
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
	selectedCourse := maxCourses[rand.Intn(len(maxCourses))]
	BroadcastSelectedCourse(selectedCourse)
}

// Player has finished loading
func Loaded(conn net.Conn) {
	GLobby.Clients[conn].Loaded = true

	for _, client := range GLobby.Clients {
		// Return if not all clients are done loading. Including observers
		if (client.Loaded == false) {
			return
		}
	}
	// Start Game
	BroadcastPacket(LoadedPacket)
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
		message := fmt.Sprintf("%s left, freeing slot %d", client.Username, client.Slot + 1)
		GLobby.VacantSlots = append(GLobby.VacantSlots, client.Slot)
		GLobby.PlayerCount--
		//fmt.Printf("%s left, freeing slot %d\n", client.Username, client.Slot + 1)
		BroadcastStringTCP(message);
	} else {
		message := fmt.Sprintf("Observer %s left", client.Username)
		BroadcastStringTCP(message);
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
	msg := fmt.Sprintf("%s: %s", client.Username, message);
	BroadcastStringTCP(msg);
}

