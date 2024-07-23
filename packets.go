package main

import (
	"fmt"
	"bufio"
	"net"
	"bytes"
	"encoding/binary"
	"math/rand"
)

const (
	// TCP Packets
	JoinPacket = iota
	LeavePacket
	MessagePacket
	LoadedPacket
	ReadyUpPacket
	SetCharacterPacket
	CoursePacket
	AssignPlayerSlotsPacket
	StartSessionPacket
	// UDP Packets
	PlayerPacket
	ActorPacket
	ObjectPacket
)

type Packet struct {
	Type    uint8
	Payload []byte
}

// Fill the rest of the player slots with AI
func AddAI(nClients []NetworkClient) []NetworkClient {

	fmt.Printf("Human Players: %d\n", GLobby.PlayerCount)
	
	for GLobby.PlayerCount < 8 {
		var slot int
		if len(GLobby.VacantSlots) > 0 {
			// Assign the first available slot
			slot = GLobby.VacantSlots[0]
			GLobby.VacantSlots = GLobby.VacantSlots[1:]
		} else {
			// Assign a new slot
			slot = GLobby.PlayerCount
			GLobby.PlayerCount++
		}
		
		randomNumber := rand.Intn(8)
		// Not a real connection. Could use the first connected player as the computer to control AI
		nClient := NetworkClient{
			Username:  getRandomUsername(),
			Slot:      slot,
			IsPlayer:  true,
			IsAI: true,
			Character: randomNumber,
			HasAuthority: false,
		}
		nClients = append(nClients, nClient)
	}
	return nClients
}

func BroadcastPlayerSlots() {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	// Generate the list of network clients (discludes observers)
	var nClients []NetworkClient
	for _, client := range GLobby.Clients {
		if (client.IsPlayer) {
			nClient := NetworkClient{
				Username:  client.Username,
				Slot:      client.Slot,
				IsPlayer:  client.IsPlayer,
				IsAI: false,
				Character: client.Character,
			}
			nClients = append(nClients, nClient)
			client.NetClient = nClient
		}
	}

	// Fill remaining slots with AI players
	nClients = AddAI(nClients)

	for _, client := range nClients {
		fmt.Printf("Slot %d\n",client.Slot)
	}

	for conn, client := range GLobby.Clients {
		var buffer bytes.Buffer
		for _, nClient := range nClients {
			// Set HasAuthority if the client matches the current connection
			if client.NetClient == nClient {
				nClient.HasAuthority = true
			} else {
				nClient.HasAuthority = false
			}

			binary.Write(&buffer, binary.LittleEndian, int32(len(nClient.Username)))
			buffer.Write([]byte(nClient.Username))
			binary.Write(&buffer, binary.LittleEndian, int32(nClient.Slot))
			binary.Write(&buffer, binary.LittleEndian, boolToInt(nClient.IsPlayer))
			binary.Write(&buffer, binary.LittleEndian, boolToInt(nClient.IsAI))
			binary.Write(&buffer, binary.LittleEndian, int32(nClient.Character))
			binary.Write(&buffer, binary.LittleEndian, boolToInt(nClient.HasAuthority))
		}

		data := buffer.Bytes()
		SendBinaryTCP(conn, AssignPlayerSlotsPacket, data)
	}
}

func SendBinaryTCP(conn net.Conn, packetType uint8, data []byte) {
	
	// Format packet: type:data
	formattedPacket := formatPacketBytesTCP(packetType, data)
	
	writer := bufio.NewWriter(conn)
	_, err := writer.Write(formattedPacket)
	if err != nil {
		fmt.Println("Error writing to client:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer:", err)
	}
	fmt.Println("Sent binary data to client:", conn.RemoteAddr())
}

func BroadcastSelectedCourse(selectedCourse int) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	data := make([]byte, 4)
    binary.LittleEndian.PutUint32(data, uint32(selectedCourse))

	packet := formatPacketBytesTCP(CoursePacket, data)

	//message := fmt.Sprintf("%d", selectedCourse)
	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		_, err := writer.Write(packet)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
		}
	}
	fmt.Printf("Selected course: %d\n", selectedCourse)
}

// func BroadcastMessageTCP(caller *Client, message string) {
// 	data := []byte(message)
// 	packet := formatPacketStringTCP(MessagePacket, data)

// 	for _, client := range GLobby.Clients {
// 		writer := bufio.NewWriter(client.Conn)
// 		_, err := writer.Write(packet)
// 		if err != nil {
// 			fmt.Println("Error writing to client:", err)
// 			continue
// 		}
// 		err = writer.Flush()
// 		if err != nil {
// 			fmt.Println("Error flushing buffer:", err)
// 		}
// 	}
// 	fmt.Println(message)
// }

func BroadcastStringTCP(message string) {
    data := []byte(message)
    packet := formatPacketStringTCP(MessagePacket, data)

	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		
		_, err := writer.Write(packet)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
		}
	}
	fmt.Printf("Broadcast: %s\n", message)
}

func BroadcastPacket(packetType uint8) {
    packet := formatPacketBytesTCP(packetType, []byte{0})

	for _, client := range GLobby.Clients {
		writer := bufio.NewWriter(client.Conn)
		_, err := writer.Write(packet)
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

func SendMessageToPlayer(client *Client, message string) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	data := []byte(message)
    packet := formatPacketStringTCP(MessagePacket, data)

	writer := bufio.NewWriter(client.Conn)
	
	_, err := writer.Write(packet)
	if err != nil {
		fmt.Println("Error writing to client:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer:", err)
	}

	fmt.Printf("Sent message to %s, %s\n", client.Username, message)
}