package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

// Start server
func main() {
	rand.Seed(time.Now().UnixNano())
	// Start TCP server
	go startTCPServer(":64010")

	// Start UDP server
	go startUDPServer(":64011") // Different port for UDP

	// Prevent main from exiting
	select {}
}

func startTCPServer(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("TCP server started on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		go HandleTCPConnection(conn)
	}
}

func startUDPServer(port string) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("UDP server started on port", port)

	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}
		go HandleUDPConnection(conn, addr, buffer[:n])
	}
}

func HandleTCPConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 4096) // Buffer to hold incoming data
	fmt.Println("Connection from client")
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client")
			} else {
				fmt.Println("Error reading from connection:", err)
			}
			return
		}

		data := buffer[:n]
		processTLVData(conn, data)
	}
}

// processTLVData processes data received in TLV format.
func processTLVData(conn net.Conn, data []byte) {
	index := 0
	length := len(data)

	for index < length {
		if index+2 > length {
			fmt.Println("Incomplete TLV data")
			return
		}

		// Read the type and length
		tlvType := data[index]
		tlvLength := int(uint16(data[index+1]))
		index += 3

		// Check if we have enough data for the value
		if index+tlvLength > length {
			fmt.Printf("type: %d, tlvLength: %d, length: %d, index: %d\n", tlvType, tlvLength, length, index)
			fmt.Println("Incomplete TLV value")
			return
		}

		// Extract the value
		value := data[index : index+tlvLength]
		index += tlvLength

		// Process the TLV
		handleTLV(conn, tlvType, value)
	}
}

// handleTLV processes each TLV entry.
func handleTLV(conn net.Conn, packetType uint8, value []byte) {
	switch packetType {
	case JoinPacket:
		Join(conn, string(value))
	case LeavePacket:
		Leave(conn)
	case MessagePacket:
		Message(conn, string(value))
	case LoadedPacket:
		Loaded(conn)
	case ReadyUpPacket:
		ReadyUp(conn, value)
	case SetCharacterPacket:
		SetCharacter(conn, value)
	case CupVotePacket:
		CupVote(conn, value)
	default:
		fmt.Println("Unknown packet type received: ", packetType)
	}
}

// Function to find a client by UDP address
// func FindClientByUDPAddr(searchAddr *net.UDPAddr) (*Client, bool) {
// 	GLobby.Mutex.Lock()
// 	defer GLobby.Mutex.Unlock()

// 	for _, client := range GLobby.Clients {
// 		fmt.Print(client.Conn.RemoteAddr().String())
// 		fmt.Print(searchAddr.String())
// 		if client.UDPConn == searchAddr.String() {
// 			return client, true
// 		}
// 	}
// 	// Client not found
// 	return nil, false
// }

func HandleUDPConnection(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	//fmt.Printf("Received UDP message from %s: 0x%X\n", addr, data)

	// Make sure the UDP client has already done a TCP handshake.
	// client, ok := FindClientByUDPAddr(addr)
	// if !ok {
	// 	//fmt.Printf("Couldn't find client with UDP address %s\n", addr);
	// 	return
	// }

	packet := ParsePacket(data)

	id := packet.Id

	client := GLobby.UUIDs[id]

	if client == nil {
		fmt.Printf("client is nil %s\n", id)
		return
	}

	fmt.Printf("packet type %d %s\n", packet.Type, id)
	switch packet.Type {
	case RegisterUDPPacket:
		RegisterConnectionUDP(conn, addr, id)
	case PlayerPacket:
		ReplicatePlayer(client, conn, packet.Payload)
	case ActorPacket:
		ReplicateActor(client, conn, packet.Payload)
	case ObjectPacket:
		ReplicateObject(client, conn, packet.Payload)
	default:
		fmt.Println("Unknown UDP packet type received: ", packet.Type)
	}
}
