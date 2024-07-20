package main

const (
	JoinPacket = iota
	LeavePacket
	MessagePacket
	LoadedPacket
	PlayerPacket
	ActorPacket
	ObjectPacket
)

type Packet struct {
	Type    int
	Payload string
}