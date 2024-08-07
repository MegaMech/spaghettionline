package main

import (
	"net"
)

type Vec3s [3]int16
type Vec3f [3]float32

type Player struct {
    /* 0x0000 */ typea uint16; // playerType?
    /* 0x0002 */ unk_002 uint16;
    /* 0x0004 */ currentRank int16;
    /* 0x0006 */ unk_006 uint16;
    /* 0x0008 */ lapCount int16;
    /* 0x000A */ unk_00A [2]int8;
    /* 0x000C */ soundEffects int32; // Bitflag.
    /* 0x0010 */ currentItemCopy int16; // Has no effect on what item the players has, It is just a synced copy
    /* 0x0012 */ unk_012 int16;
	/* 0x0014 */ pos Vec3f;
    /* 0x0020 */ copy_rotation_x float32;
    /* 0x0024 */ copy_rotation_y float32;
    /* 0x0028 */ copy_rotation_z float32;
    /* 0x002C */ rotation Vec3s;
    /* 0x0032 */ unk_032[0x2] [2]int8;
    /* 0x0034 */ velocity Vec3f;
}

func ReplicatePlayer(client *Client, conn *net.UDPConn, data []byte) {
	// repData, err := parseReplicationData(data)
	// if err != nil {
	// 	fmt.Println("Error parsing replication data:", err)
	// 	return
	// }
	// packetData, err := serializeReplicationData(repData)
	// if err != nil {
	// 	fmt.Println("Error serializing replication data:", err)
	// 	return
	// }

	ReplicationBroadcast(client, PlayerPacket, data);
}

func ReplicateActor(client *Client, conn *net.UDPConn, data []byte) {

}

func ReplicateObject(client *Client, conn *net.UDPConn, data []byte) {

}