package main

import (
	"fmt"
	"net"
	"bufio"
	"strings"
)

// Start server
func main() {
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
		go HandleTCPMessage(conn)
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

	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}
		go HandleUDPMessage(conn, addr, buffer[:n])
	}
}

func HandleTCPMessage(conn net.Conn) {
	defer conn.Close()

	fmt.Println("New connection established!")

	reader := bufio.NewReader(conn)

	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		data = strings.TrimSpace(data) // Remove trailing newline
		//DebugPacket(data)
		packet := ParsePacket(data)
		switch packet.Type {
			case JoinPacket:
				Join(conn, packet.Payload);
			case LeavePacket:
				Leave(conn);
			case MessagePacket:
				Message(conn, packet.Payload);
			default:
				fmt.Println("Unknown packet type received: ", packet.Type)
		}
	}
}

// Function to find a client by UDP address
func FindClientByUDPAddr(searchAddr *net.UDPAddr) (*Client, bool) {
	GLobby.Mutex.Lock()
	defer GLobby.Mutex.Unlock()

	for _, client := range GLobby.Clients {
		if client.Conn.RemoteAddr().String() == searchAddr.String() {
			return client, true
		}
	}
	// Client not found
	return nil, false
}

func HandleUDPMessage(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	fmt.Printf("Received UDP message from %s: %s\n", addr, data)

	// Make sure the UDP client has already done a TCP handshake.
	client, ok := FindClientByUDPAddr(addr)
	if !ok {
		return
	}

	packet := ParsePacket(string(data))
	switch packet.Type {
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

func ParsePacket(data string) Packet {
	// Split the data into two parts: type and payload
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		// Return an invalid packet if the split was unsuccessful
		return Packet{Type: -1}
	}

	var packetType int
	_, err := fmt.Sscanf(parts[0], "%d", &packetType)
	if err != nil {
		return Packet{Type: -1} // Invalid packet type
	}
	return Packet{Type: packetType, Payload: parts[1]}
}

func DebugPacket(data string) {
	fmt.Println("Raw packet data:", data)

	// Call ParsePacket to see how it parses the data
	packet := ParsePacket(data)
	fmt.Printf("Parsed Packet Type: %d\n", packet.Type)
	fmt.Printf("Parsed Packet Payload: %s\n", packet.Payload)
}