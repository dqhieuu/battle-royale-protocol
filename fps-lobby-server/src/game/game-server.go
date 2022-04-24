package game

import "net"

type GameServer struct {
	Address         string
	PacketFrequency int
	Connection      net.Conn
}

func NewGameServer(address string, packetFreq int, connection net.Conn) *GameServer {
	return &GameServer{
		Address:         address,
		PacketFrequency: packetFreq,
		Connection:      connection,
	}
}
