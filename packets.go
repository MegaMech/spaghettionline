package main

const (
	JoinPacket = iota
	LeavePacket
	MessagePacket
	PlayerPacket
	ActorPacket
	ObjectPacket
)

type Packet struct {
	Type    int
	Payload string
}