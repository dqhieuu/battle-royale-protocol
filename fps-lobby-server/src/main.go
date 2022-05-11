package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"fps-lobby-server/src/game"
	"fps-lobby-server/src/utils"
	g "github.com/AllenDang/giu"
	"github.com/xtaci/kcp-go"
	"image/color"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var OutboundIp = utils.GetOutboundIP()
var PublicIp *string

var AccountDatabase sync.Map // a map of username -> *Account

var gameServers []*game.GameServer
var AddressToGameServer sync.Map // a map of address -> *GameServer

func loop() {
	var gameServerRows []*g.TableRowWidget

	for i := range gameServers {
		gameServerRows = append(gameServerRows, g.TableRow(
			g.Label(gameServers[i].Address),
			g.Label(strconv.Itoa(gameServers[i].PacketFrequency)),
		))
	}

	g.SingleWindow().Layout(
		g.Style().SetFontSize(32).To(g.Label("Server lobby cho game PWNBG")),
		g.Style().
			SetColor(g.StyleColorText, color.RGBA{100, 255, 255, 255}).
			To(g.Label(fmt.Sprintf("IP: %s",
				OutboundIp,
			))),

		g.Table().Columns(
			g.TableColumn("Địa chỉ server game"),
			g.TableColumn("Tần suất packet(Hz)"),
		).Rows(
			gameServerRows...,
		).Freeze(1, 0).Size(400, 100),
		g.Row(
			g.Label("Scale server game"),
			g.Button("Up").OnClick(func() {
				scaleGameServer(true)
			}),
			g.Button("Down").OnClick(func() {
				scaleGameServer(false)
			}),
		),
	)
	g.Update()
}

func HttpListener(w http.ResponseWriter, r *http.Request) {
	var jsonMap map[string]any
	err := json.NewDecoder(r.Body).Decode(&jsonMap)
	if err != nil {
		log.Println(err)
		return
	}

	badRequest := false
	requestType := jsonMap["alert"].(string)
	switch requestType {
	case "scale-up":
		scaleGameServer(true)
	case "scale-down":
		scaleGameServer(false)
	default:
		badRequest = true
	}

	if badRequest {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func scaleGameServer(isScaleUp bool) {
	var buf bytes.Buffer
	utils.WriteAs1Byte(&buf, utils.PacketType_ScaleAlertToGameServer)
	if isScaleUp {
		log.Println("Scaling up game servers")
		utils.WriteAs1Byte(&buf, utils.PacketSubType_ScaleAlertToGameServer_ScaleUp)
	} else {
		log.Println("Scaling down game servers")
		utils.WriteAs1Byte(&buf, utils.PacketSubType_ScaleAlertToGameServer_ScaleDown)
	}
	packetToSend := buf.Bytes()

	for i := range gameServers {
		gameServers[i].Connection.Write(packetToSend)
	}
}

func handleGameServer(conn net.Conn) {
	defer conn.Close()
	var currentGameServer *game.GameServer
	for {
		if currentGameServer != nil && currentGameServer.Connection != conn {
			// another game server is already connected
			return
		}
		buffer := make([]byte, 1024)
		if _, err := conn.Read(buffer); err == nil {
			var byteBuffer bytes.Buffer
			byteBuffer.Write(buffer)
			packetType := utils.ReadAs1Byte(&byteBuffer)

			switch packetType {
			case utils.PacketType_GameServerConnectToLobbyServer:
				packetSubType := utils.ReadAs1Byte(&byteBuffer)
				switch packetSubType {
				case utils.PacketSubType_GameServerConnectToLobbyServer_AddToLobby:
					gameServerIp := conn.RemoteAddr().(*net.UDPAddr).IP.String()
					serverPacketFreq := utils.ReadAs2Byte(&byteBuffer)

					existingGameServer, _ := AddressToGameServer.Load(gameServerIp)
					if existingGameServer == nil {
						currentGameServer = game.NewGameServer(gameServerIp, serverPacketFreq, conn)
						gameServers = append(gameServers, currentGameServer)
						AddressToGameServer.Store(gameServerIp, currentGameServer)
					} else {
						currentGameServer = existingGameServer.(*game.GameServer)
						currentGameServer.Connection = conn
						currentGameServer.PacketFrequency = serverPacketFreq
					}

					byteBuffer.Reset()
					utils.WriteAs1Byte(&byteBuffer, utils.PacketType_GameServerConnectToLobbyServer)
					utils.WriteAs1Byte(&byteBuffer, utils.PacketSubType_GameServerConnectToLobbyServer_AddToLobbySuccess)
					conn.Write(byteBuffer.Bytes())
				}
			case utils.PacketType_GameServerPeriodicToLobbyServer:
				packetSubType := utils.ReadAs1Byte(&byteBuffer)
				switch packetSubType {
				case utils.PacketSubType_GameServerPeriodicToLobbyServer_RoomStatus:
					serverPacketFreq := utils.ReadAs2Byte(&byteBuffer)
					// there are many more data here, but i will ignore them for now
					currentGameServer.PacketFrequency = serverPacketFreq
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	if _, err := conn.Read(buffer); err == nil {
		var byteBuffer bytes.Buffer
		byteBuffer.Write(buffer)
		packetType := utils.ReadAs1Byte(&byteBuffer)

		switch packetType {
		case utils.PacketType_ClientConnectToLobbyServer:
			packetSubType := utils.ReadAs1Byte(&byteBuffer)
			switch packetSubType {
			case utils.PacketSubType_ClientConnectToLobbyServer_LoginRequest:
				username := utils.ReadString(&byteBuffer)

				byteBuffer.Reset()
				utils.WriteAs1Byte(&byteBuffer, utils.PacketType_ClientConnectToLobbyServer)
				utils.WriteAs1Byte(&byteBuffer, utils.PacketSubType_ClientConnectToLobbyServer_LoginSuccess)
				conn.Write(byteBuffer.Bytes())

				account, _ := AccountDatabase.Load(username)
				if account == nil {
					account = game.NewAccount(username)
					AccountDatabase.Store(username, account)
				}
				var gameServerAddress = account.(*game.Account).GameServerAddress

				if gameServerAddress == nil {
					for {
						if len(gameServers) > 0 {
							randGameServer := rand.Intn(len(gameServers))
							gameServerAddress = &gameServers[randGameServer].Address
							account.(*game.Account).GameServerAddress = gameServerAddress
							break
						}
						time.Sleep(time.Millisecond * 100)
					}
				}

				if utils.IsPrivateIp(gameServerAddress) && PublicIp != nil {
					gameServerAddress = PublicIp
				}

				byteBuffer.Reset()
				utils.WriteAs1Byte(&byteBuffer, utils.PacketType_LobbyServerToClient)
				utils.WriteAs1Byte(&byteBuffer, utils.PacketSubType_LobbyServerToClient_GameServerAddress)
				utils.WriteString(&byteBuffer, *gameServerAddress)
				conn.Write(byteBuffer.Bytes())
				conn.Read(buffer) // hacking so it doesn't close the connection
			}
		}
	} else {
		log.Fatal(err)
	}
}

func GameServerConnectionListener(url string) {
	if listener, err := kcp.Listen(url); err == nil {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleGameServer(s)
		}
	} else {
		log.Fatal(err)
	}
}

func ClientConnectionListener(url string) {
	if listener, err := kcp.Listen(url); err == nil {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleClient(s)
		}
	} else {
		log.Fatal(err)
	}
}

const GameServerToLobbyServerPort = ":19000"
const ClientToLobbyServerPort = ":19001"
const AlerterToLobbyServerPort = ":19006"

func main() {
	rand.Seed(time.Now().UnixNano())

	go func() {
		fetchedIp := utils.GetPublicIp()
		if fetchedIp != "" {
			PublicIp = &fetchedIp
		}
	}()

	go GameServerConnectionListener(GameServerToLobbyServerPort)
	go ClientConnectionListener(ClientToLobbyServerPort)

	http.HandleFunc("/", HttpListener)
	go http.ListenAndServe(AlerterToLobbyServerPort, nil)

	wnd := g.NewMasterWindow("PWNBG Lobby Server", 500, 300, 0)

	wnd.Run(loop)
}
