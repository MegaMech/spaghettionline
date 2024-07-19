package main

import (
	"fmt"
	"net"
)

func ReplicationBroadcast(caller *Client, packetType int, packetData []byte) {
	// Broadcast data to all connected clients
	for _, client := range GLobby.Clients {
		if client == caller {
			continue // Skip the calling client
		}
		// Cast the connection to *net.UDPConn
		if udpConn, ok := client.Conn.(*net.UDPConn); ok {
			// Send the packet data to the client's address
			_, err := udpConn.WriteToUDP(packetData, client.Conn.RemoteAddr().(*net.UDPAddr))
			if err != nil {
				fmt.Println("Error sending data to client:", err)
			}
		} else {
			fmt.Println("Client connection is not of type *net.UDPConn")
		}
	}
}

func parseReplicationData(data string) (string, bool) {
	return "data", true
}

func serializeReplicationData(data string) (string, bool) {
	return "data", true
}