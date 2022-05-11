package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
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

type IP struct {
	Query string
}

func GetPublicIp() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	var ip IP
	json.Unmarshal(body, &ip)

	return ip.Query
}

func IsPrivateIp(ip *string) bool {
	if ip == nil {
		return false
	}
	ipAddress := net.ParseIP(*ip)
	return ipAddress.IsPrivate()
}
