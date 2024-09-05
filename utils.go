package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"regexp"

	"github.com/google/uuid"
)

func ParsePacket(data []byte) *Packet {
	if len(data) < 3 {
		fmt.Print("Not enough UDP data to read type and length")
		return nil
	}

	packetType := data[0]

	myUUID, err := uuid.FromBytes(data[1:17])
	if err != nil {
		fmt.Println("Error converting byte array to UUID:", err)
		return nil
	}
	payloadLength := uint16(data[18]) | uint16(data[19])<<8

	// Not reading payloadLenght correctly
	// if len(data) < int(3+payloadLength) {
	// 	fmt.Printf("Not enough data for the whole UDP packet: 0x%X data len: 0x%X\n", payloadLength, len(data))
	//     fmt.Printf("type: 0x%d\n", packetType)
	// 	return nil
	// }

	packetData := data[20 : payloadLength+20]
	return &Packet{Type: packetType, Id: myUUID, Size: payloadLength, Payload: packetData}

	//return packet
}

// func DebugPacket(data []byte) {
// 	fmt.Println("Raw packet data:", data)

// 	// Call ParsePacket to see how it parses the data
// 	packet := ParsePacket(data)
// 	fmt.Printf("Parsed Packet Type: %d\n", packet.Type)
// 	fmt.Printf("Parsed Packet Payload: %s\n", packet.Payload)
// }

// Allows letters and numbers only
func isValidUsername(username string) bool {
	//fmt.Printf("Debug Username: %s\n", username)
	//fmt.Printf("Debug Username: %X\n", username)
	re := regexp.MustCompile("^[a-zA-Z0-9]+$")
	return re.MatchString(username)
}

func ReplicationBroadcastUDP(caller *Client, packetType int, packetData []byte) {
	// Broadcast data to all connected clients
	for _, client := range GLobby.ClientsUDP {
		if client == caller {
			continue // Skip the calling client
		}

		// Send the packet data to the client's address
		_, err := client.ConnUDP.WriteToUDP(packetData, client.addr)
		if err != nil {
			fmt.Println("Error sending data to client:", err)
		}
	}
}

func parseReplicationData(data string) (string, bool) {
	return "data", true
}

func serializeReplicationData(data string) (string, bool) {
	return "data", true
}

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// func formatPacketStringTCP(packetType int, payload string) string {
//     return fmt.Sprintf("%d:%s\n", packetType, payload)
// }

func formatPacketBytesTCP(packetType uint8, data []byte) []byte {
	// Determine the length of the value
	dataLength := uint16(len(data))

	// Create a buffer to hold the entire packet
	// 1 byte for type + 2 bytes for length + length of data
	packet := make([]byte, 1+2+dataLength)

	// Set the packet type
	packet[0] = packetType

	// Set the length of the data
	// Note: We use a little-endian encoding for length here
	packet[1] = byte(dataLength)
	packet[2] = byte(dataLength >> 8)

	// Copy the data into the packet
	copy(packet[3:], data)

	return packet
}

func formatPacketStringTCP(packetType uint8, data []byte) []byte {
	// Determine the length of the value
	dataLength := uint16(len(data))

	// Create a buffer to hold the entire packet
	// 1 byte for type + 2 bytes for length + length of data + 1 byte for string terminator character
	packet := make([]byte, 1+2+dataLength+1)

	// Set the packet type
	packet[0] = packetType

	// Set the length of the data
	// Note: We use a little-endian encoding for length here
	packet[1] = byte(dataLength)
	packet[2] = byte(dataLength >> 8)

	// Copy the data into the packet
	copy(packet[3:], data)

	// Add the string terminator
	packet[3+dataLength] = 0

	return packet
}

// func formatPacketBytesTCP(packetType int, data []byte) []byte {
// 	packetTypeByte := byte(packetType)
// 	packetTypeBytes := []byte{packetTypeByte, ':'}
// 	return append(packetTypeBytes, data...)
// }

var coolUsernames = []string{
	"ShadowHunter",
	"LunarEclipse",
	"CyberNinja",
	"NebulaRider",
	"QuantumStorm",
	"PhantomGamer",
	"VortexDynamo",
	"MysticVoyager",
	"StellarFury",
	"EclipseSpecter",
	"NovaGuardian",
	"InfernoStriker",
	"AstralWanderer",
	"CelestialKnight",
	"TitanBlaze",
	"RogueSpecter",
	"GalacticSavior",
	"ZenithPhoenix",
	"ThunderStrike",
	"FrostbiteLegend",
}

func getRandomUsername() string {
	index := rand.Intn(len(coolUsernames))
	name := coolUsernames[index]
	coolUsernames = append(coolUsernames[:index], coolUsernames[index+1:]...)
	return name
}

func uint32ToIP(ipUint32 uint32) net.IP {
	ipBytes := make([]byte, 4)

	// Assuming the input is in little-endian format, we need to use LittleEndian
	binary.LittleEndian.PutUint32(ipBytes, ipUint32)

	// Reverse the byte slice because net.IP expects big-endian for proper IP formatting
	return net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
}
