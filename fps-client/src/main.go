package main

import (
	"bytes"
	_ "embed"
	"fmt"
	game "fps-client/src/client"
	"fps-client/src/utils"
	g "github.com/AllenDang/giu"
	"github.com/golang/geo/r2"
	"github.com/xtaci/kcp-go"
	"image/color"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var clientsToAdd int32 = 30
var currentConnections int32 = 0
var lobbyAddr = "127.0.0.1"
var lobbyPort = ":19003"
var isFrozen = false

var currentConnectionsMutex = &sync.Mutex{}

func loop() {
	currentConnectionsMutex.Lock()
	currentConnections := currentConnections
	currentConnectionsMutex.Unlock()

	var frozenButtonText string
	var frozenConnectionText string
	if isFrozen {
		frozenButtonText = "Tiếp tục gửi gói tin"
		frozenConnectionText = " (Đang tạm dừng)"
	} else {
		frozenButtonText = "Tạm dừng gửi gói tin"
	}

	g.SingleWindow().Layout(
		g.Style().SetFontSize(32).To(g.Label("Client spawner cho server PWNBG")),
		g.Row(
			g.Label("Địa chỉ IP game lobby"),
			g.InputText(&lobbyAddr).Size(200),
		),

		g.Row(
			g.Label("Số lượng client thêm vào"),
			g.InputInt(&clientsToAdd).Size(60),
			g.Button("Thêm client mới").OnClick(func() {
				for i := 0; i < int(clientsToAdd); i++ {
					go clientToGameServer(lobbyAddr + lobbyPort)
				}
			}),
		),
		g.Style().
			SetColor(g.StyleColorText, color.RGBA{100, 255, 255, 255}).
			To(g.Label("Số kết nối hiện tại: "+strconv.Itoa(int(currentConnections))+frozenConnectionText)),
		g.Button(frozenButtonText).OnClick(func() {
			isFrozen = !isFrozen
		}),
	)
}

func updatePlayerDataToServer(state *game.RoomState, player *game.Player, packetFreq *int, conn net.Conn) {
	if state == nil {
		return
	}
	currentConnectionsMutex.Lock()
	currentConnections++
	currentConnectionsMutex.Unlock()

Loop:
	for {
		switch *state {
		case game.FINISHED:
			break Loop
		case game.WAITING:
			time.Sleep(time.Millisecond * 100)
			continue
		case game.PLAYING:
			if isFrozen {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			if player.IsDead() {
				break Loop
			}
			player.UpdateLocation()
			var buf bytes.Buffer
			utils.WriteAs1Byte(&buf, utils.PacketType_GameServerAndClientInGame)
			utils.WriteAs1Byte(&buf, utils.PacketSubType_GameServerAndClientInGame_PositionUpdate)
			utils.WriteAs2Byte(&buf, int(player.Location.X))
			utils.WriteAs2Byte(&buf, int(player.Location.Y))
			conn.Write(buf.Bytes())
			sleepTime := time.Millisecond * time.Duration(1000.0/float64(*packetFreq))
			time.Sleep(sleepTime)
		}
	}

	currentConnectionsMutex.Lock()
	currentConnections--
	currentConnectionsMutex.Unlock()

}

func clientToGameServer(url string) {
	time.Sleep(time.Second)

	playerName := fmt.Sprintf("player_%02x", rand.Int31())

	if sess, err := kcp.Dial(url); err == nil {
		buffer := make([]byte, 1024)

		var buf bytes.Buffer
		utils.WriteAs1Byte(&buf, utils.PacketType_ClientConnectToGameServer)
		utils.WriteAs1Byte(&buf, utils.PacketSubType_ClientConnectToGameServer_JoinRoom)
		utils.WriteString(&buf, playerName)
		sess.Write(buf.Bytes())

		buffer = make([]byte, 1024)
		if _, err := sess.Read(buffer); err == nil {
			var byteBuffer bytes.Buffer
			byteBuffer.Write(buffer)
			packetType := utils.ReadAs1Byte(&byteBuffer)
			if packetType != utils.PacketType_ClientConnectToGameServer {
				log.Fatal("Wrong packet type")
				return
			}
			packetSubType := utils.ReadAs1Byte(&byteBuffer)
			if packetSubType != utils.PacketSubType_ClientConnectToGameServer_JoinRoomSuccess {
				log.Fatal("Wrong packet sub type")
				return
			}
			hasRoomData := utils.ReadAs1Byte(&byteBuffer)
			if hasRoomData == 0 {
				log.Fatal("No room data")
				return
			}

			posX := utils.ReadAs2Byte(&byteBuffer)
			posY := utils.ReadAs2Byte(&byteBuffer)
			HP := utils.ReadAs2Byte(&byteBuffer)
			gunId := utils.ReadAs2Byte(&byteBuffer)
			hatId := utils.ReadAs2Byte(&byteBuffer)
			shirtId := utils.ReadAs2Byte(&byteBuffer)
			pantsId := utils.ReadAs2Byte(&byteBuffer)
			packetFreq := utils.ReadAs2Byte(&byteBuffer)
			roomState := game.RoomState(utils.ReadAs1Byte(&byteBuffer))
			mapWidth := utils.ReadAs2Byte(&byteBuffer)
			mapHeight := utils.ReadAs2Byte(&byteBuffer)

			gameMap := game.NewGameMap(mapWidth, mapHeight)

			player := game.Player{
				HP:   HP,
				Name: playerName,
				Inventory: game.Inventory{
					Gun:   gameMap.ItemStore.QueryItem(gunId),
					Hat:   gameMap.ItemStore.QueryItem(hatId),
					Shirt: gameMap.ItemStore.QueryItem(shirtId),
					Pants: gameMap.ItemStore.QueryItem(pantsId),
				},
				Location:  r2.Point{float64(posX), float64(posY)},
				Direction: r2.Point{0, 1},
				IsBot:     true,
			}
			player.SetMap(gameMap)
			player.Location = r2.Point{float64(posX), float64(posY)} // coding problem will fix later

			go updatePlayerDataToServer(&roomState, &player, &packetFreq, sess)
			for {
				buffer = make([]byte, 1024)
				if _, err := sess.Read(buffer); err == nil {
					var byteBuffer bytes.Buffer
					byteBuffer.Write(buffer)
					packetType := utils.ReadAs1Byte(&byteBuffer)
					if packetType != utils.PacketType_GameServerAndClientInGame {
						log.Fatal("Wrong packet type 2")
						return
					}
					packetSubType := utils.ReadAs1Byte(&byteBuffer)
					switch packetSubType {
					case utils.PacketSubType_GameServerAndClientInGame_HPUpdate:
						currentHp := utils.ReadAs2Byte(&byteBuffer)
						player.HP = currentHp
						if player.IsDead() {
							return
						}
					case utils.PacketSubType_GameServerAndClientInGame_StateUpdate:
						roomState = game.RoomState(utils.ReadAs1Byte(&byteBuffer))
						if roomState == game.FINISHED {
							return
						}
					}
				}
			}
		} else {
			log.Fatal(err)
		}

	} else {
		log.Println(err)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	wnd := g.NewMasterWindow("PWNBG Client Spawner", 500, 300, 0)

	wnd.Run(loop)
}
