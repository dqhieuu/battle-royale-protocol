package utils

import (
	"log"
	"math/rand"
	"net"
)

func RandRangeInt(min, max int) int {
	if min == max {
		return min
	}
	return min + rand.Intn(max-min)
}

func RandRangeFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
