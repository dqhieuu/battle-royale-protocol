package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"fps-game-server/src/game"
	"fps-game-server/src/utils"
	g "github.com/AllenDang/giu"
	"github.com/paulbellamy/ratecounter"
	"github.com/xtaci/kcp-go"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var texture *g.Texture

var selectedRoomIndex int
var selectedPlayerIndex = -1
var playerRoomLookupByName sync.Map
var playerSessionLookupByName sync.Map
var rooms []*game.GameRoom

var packetFrequency = 64

const PacketFrequencyUpperLimit = 256
const PacketFrequencyLowerLimit = 1

var packetCounter = ratecounter.NewRateCounter(1 * time.Second)

func addPlayerByName(name string, isBot bool) *game.Player {
	searchedRoom, _ := playerRoomLookupByName.Load(name)
	if searchedRoom != nil {
		return searchedRoom.(*game.GameRoom).PlayerByName(name)
	}

	hasEmptyRoom := false
	var emptyRoom *game.GameRoom
	for i := range rooms {
		if !rooms[i].IsFull() && rooms[i].State == game.WAITING {
			hasEmptyRoom = true
			emptyRoom = rooms[i]
			break
		}
	}

	if !hasEmptyRoom {
		emptyRoom = game.NewGameRoom()
		rooms = append(rooms, emptyRoom)
		go emptyRoom.Activate()
	}

	var newPlayer *game.Player
	if isBot {
		newPlayer = game.NewBotPlayerWithName(name)
	} else {
		newPlayer = game.NewPlayerWithName(name)
	}
	newPlayer.SetMap(emptyRoom.Map)
	emptyRoom.AddPlayer(newPlayer)
	playerRoomLookupByName.Store(name, emptyRoom)

	return newPlayer
}

func scaleFrequencyUp() {
	if packetFrequency >= PacketFrequencyUpperLimit {
		return
	}
	packetFrequency *= 2
}

func scaleFrequencyDown() {
	if packetFrequency <= PacketFrequencyLowerLimit {
		return
	}
	packetFrequency /= 2
}

var mapScale float32 = 0.8

func loop() {
	var lobbyTableRows []*g.TableRowWidget

	var roomTableRows []*g.TableRowWidget

	for i := range rooms {
		i := i

		var roomState string
		switch rooms[i].State {
		case game.FINISHED:
			roomState = "Đã kết thúc"
		case game.WAITING:
			roomState = "Đang chờ"
		case game.PLAYING:
			roomState = "Đang chơi"
		}

		lobbyTableRows = append(lobbyTableRows, g.TableRow(
			g.Selectable(strconv.Itoa(i+1)).Flags(g.SelectableFlagsSpanAllColumns).OnClick(func() {
				fmt.Printf("Selected room %d\n", i+1)
				if selectedRoomIndex != i {
					selectedRoomIndex = i
					selectedPlayerIndex = -1
				}
			}),
			g.Label(fmt.Sprintf("%d/%d", rooms[i].PlayerCount(), rooms[i].MaxPlayers)),
			g.Label(roomState),
		))
	}

	var players []*game.Player
	if selectedRoomIndex >= 0 && selectedRoomIndex < len(rooms) {
		players = rooms[selectedRoomIndex].Players
	}

	for i := range players {
		var currentGun, currentHat, currentShirt, currentPants string
		if players[i].Inventory.Gun != nil {
			currentGun = players[i].Inventory.Gun.Name
		}
		if players[i].Inventory.Hat != nil {
			currentHat = players[i].Inventory.Hat.Name
		}
		if players[i].Inventory.Shirt != nil {
			currentShirt = players[i].Inventory.Shirt.Name
		}
		if players[i].Inventory.Pants != nil {
			currentPants = players[i].Inventory.Pants.Name
		}
		i := i

		roomTableRows = append(roomTableRows,
			g.TableRow(
				g.Selectable(strconv.Itoa(i+1)).Flags(g.SelectableFlagsSpanAllColumns).OnClick(func() {
					if selectedPlayerIndex == i {
						selectedPlayerIndex = -1
					} else {
						selectedPlayerIndex = i
						fmt.Println("Selected player " + players[i].Name)
					}
				}),
				g.Label(players[i].Name),
				g.Label(strconv.Itoa(players[i].HP)),
				g.Label(currentGun),
				g.Label(currentHat),
				g.Label(currentShirt),
				g.Label(currentPants),
			),
		)
	}

	g.SingleWindow().Layout(
		g.Style().
			SetColor(g.StyleColorText, color.RGBA{100, 255, 255, 255}).
			To(g.Label(fmt.Sprintf("Số người chơi tối thiểu để bắt đầu: %d. Tần suất gửi gói tin: %dHz. Số socket IO / giây: %d",
				game.MinTriggerPlayingPlayers,
				packetFrequency,
				packetCounter.Rate(),
			))),
		g.Row(
			g.Table().Columns(
				g.TableColumn("Phòng"),
				g.TableColumn("Người chơi"),
				g.TableColumn("Trạng thái"),
			).Rows(
				lobbyTableRows...,
			).Freeze(1, 1).Size(400, 100),
			g.Column(
				g.Row(
					g.Label("Kích cỡ vẽ canvas map"),
					g.InputFloat(&mapScale).Size(70),
				),
				g.Row(
					g.Label("Scale tần suất gửi packet thủ công"),
					g.Button("+").OnClick(func() {
						scaleFrequencyUp()
					}).Size(30, 30).Disabled(packetFrequency >= PacketFrequencyUpperLimit),
					g.Button("-").OnClick(func() {
						scaleFrequencyDown()
					}).Size(30, 30).Disabled(packetFrequency <= PacketFrequencyLowerLimit),
				),
			),
		),
		g.Table().Columns(
			g.TableColumn("#"),
			g.TableColumn("Tên"),
			g.TableColumn("HP"),
			g.TableColumn("Súng"),
			g.TableColumn("Mũ"),
			g.TableColumn("Áo"),
			g.TableColumn("Quần"),
		).Freeze(1, 1).Rows(
			roomTableRows...,
		).Size(900, 200),
		g.Custom(func() {
			canvas := g.GetCanvas()
			pos := g.GetCursorScreenPos()
			red := color.RGBA{200, 75, 75, 255}
			blue := color.RGBA{37, 37, 250, 255}
			yellow := color.RGBA{252, 186, 3, 255}
			if texture != nil {
				canvas.AddImage(texture, pos, pos.Add(image.Point{int(float32(400) * mapScale), int(float32(400) * mapScale)}))
			}
			for i := range players {
				circlePos := pos.Add(image.Point{int(float32(players[i].Location.X) * mapScale), int(float32(players[i].Location.Y) * mapScale)})
				canvas.AddCircleFilled(circlePos, float32(5)*mapScale, red)
				canvas.AddCircleFilled(circlePos, float32(3)*mapScale, color.White)
				if players[i].IsDead() {
					canvas.AddCircleFilled(circlePos, float32(3)*mapScale, color.Black)
				} else {
					for j := range players {
						if i == j || players[j].IsDead() {
							continue
						}
						xd := players[i].Location.X - players[j].Location.X
						xd2 := xd * xd
						yd := players[i].Location.Y - players[j].Location.Y
						yd2 := yd * yd
						if xd2+yd2 < game.SqrMaxShootableDistance {
							canvas.AddLine(
								pos.Add(image.Point{int(float32(players[i].Location.X) * mapScale), int(float32(players[i].Location.Y) * mapScale)}),
								pos.Add(image.Point{int(float32(players[j].Location.X) * mapScale), int(float32(players[j].Location.Y) * mapScale)}),
								blue, float32(3)*mapScale)
							break
						}
					}
				}
			}

			// to prevent overlapping geometry
			for i := range players {
				if i == selectedPlayerIndex {
					circlePos := pos.Add(image.Point{int(float32(players[i].Location.X) * mapScale), int(float32(players[i].Location.Y) * mapScale)})
					canvas.AddCircle(circlePos, float32(float32(20)*mapScale), yellow, 20, float32(3)*mapScale)
				}
			}

			g.Update()
		}),
	)
}

func clientConnectionListener(url string) {
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

func sendGameSnapshotToClient(conn net.Conn, player *game.Player, room *game.GameRoom, sendFreq int) {
	time.Sleep(100 * time.Millisecond) // wait for client to update
	for {
		if sendFreq != packetFrequency {
			sendFreq = packetFrequency
			var buf bytes.Buffer
			utils.WriteAs1Byte(&buf, utils.PacketType_GameServerAndClientInGame)
			utils.WriteAs1Byte(&buf, utils.PacketSubType_GameServerAndClientInGame_PacketFrequencyUpdate)
			utils.WriteAs2Byte(&buf, sendFreq)
			conn.Write(buf.Bytes())
			packetCounter.Incr(1)
		}
		var buf bytes.Buffer
		utils.WriteAs1Byte(&buf, utils.PacketType_GameServerAndClientInGame)
		utils.WriteAs1Byte(&buf, utils.PacketSubType_GameServerAndClientInGame_StateUpdate)
		utils.WriteAs1Byte(&buf, int(room.State))
		conn.Write(buf.Bytes())
		packetCounter.Incr(1)
		buf.Reset()

		utils.WriteAs1Byte(&buf, utils.PacketType_GameServerAndClientInGame)
		utils.WriteAs1Byte(&buf, utils.PacketSubType_GameServerAndClientInGame_HPUpdate)
		utils.WriteAs2Byte(&buf, player.HP)
		conn.Write(buf.Bytes())
		packetCounter.Incr(1)
		sleepTime := time.Millisecond * time.Duration(1000.0/float64(sendFreq))
		time.Sleep(sleepTime)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	var playerUsername *string
	var player *game.Player
	for {
		// check if there are 2 different clients connected as the same player
		if playerUsername != nil {
			currentMapConn, _ := playerSessionLookupByName.Load(*playerUsername)
			if currentMapConn != conn {
				break // break the loop and close the connection
			}
		}

		buffer := make([]byte, 1024)
		if _, err := conn.Read(buffer); err == nil {
			packetCounter.Incr(1) // đọc conn

			var byteBuffer bytes.Buffer
			byteBuffer.Write(buffer)
			packetType := utils.ReadAs1Byte(&byteBuffer)
			switch packetType {
			case utils.PacketType_ClientConnectToGameServer:
				packetSubType := utils.ReadAs1Byte(&byteBuffer)
				switch packetSubType {
				case utils.PacketSubType_ClientConnectToGameServer_JoinRoom:
					username := utils.ReadString(&byteBuffer)
					// used for checking if the player is already connected
					playerUsername = &username
					// we keep a reference to the player object for later use
					player = addPlayerByName(username, false)
					// sessionId is used to differentiate between different clients connected as the same player
					playerSessionLookupByName.Store(username, conn)

					byteBuffer.Reset() // reset for writing
					utils.WriteAs1Byte(&byteBuffer, utils.PacketType_ClientConnectToGameServer)
					utils.WriteAs1Byte(&byteBuffer, utils.PacketSubType_ClientConnectToGameServer_JoinRoomSuccess)
					utils.WriteAs1Byte(&byteBuffer, utils.PacketSubType_ClientConnectToLobbyServer_JoinRoomSucess_WithRoomInfo)
					utils.WriteAs2Byte(&byteBuffer, int(player.Location.X))
					utils.WriteAs2Byte(&byteBuffer, int(player.Location.Y))
					utils.WriteAs2Byte(&byteBuffer, player.HP)
					if player.Inventory.Gun != nil {
						utils.WriteAs2Byte(&byteBuffer, player.Inventory.Gun.Id)
					} else {
						utils.WriteAs2Byte(&byteBuffer, 0)
					}
					if player.Inventory.Hat != nil {
						utils.WriteAs2Byte(&byteBuffer, player.Inventory.Hat.Id)
					} else {
						utils.WriteAs2Byte(&byteBuffer, 0)
					}
					if player.Inventory.Shirt != nil {
						utils.WriteAs2Byte(&byteBuffer, player.Inventory.Shirt.Id)
					} else {
						utils.WriteAs2Byte(&byteBuffer, 0)
					}
					if player.Inventory.Pants != nil {
						utils.WriteAs2Byte(&byteBuffer, player.Inventory.Pants.Id)
					} else {
						utils.WriteAs2Byte(&byteBuffer, 0)
					}
					utils.WriteAs2Byte(&byteBuffer, packetFrequency)

					playerRoom, _ := playerRoomLookupByName.Load(username)
					utils.WriteAs1Byte(&byteBuffer, int(playerRoom.(*game.GameRoom).State))
					utils.WriteAs2Byte(&byteBuffer, game.DefaultMapWidth)
					utils.WriteAs2Byte(&byteBuffer, game.DefaultMapHeight)

					conn.Write(byteBuffer.Bytes())
					packetCounter.Incr(1) // gửi đến conn

					go sendGameSnapshotToClient(conn, player, playerRoom.(*game.GameRoom), packetFrequency)
				}
			case utils.PacketType_GameServerAndClientInGame:
				packetSubType := utils.ReadAs1Byte(&byteBuffer)
				switch packetSubType {
				case utils.PacketSubType_GameServerAndClientInGame_PositionUpdate:
					posX := utils.ReadAs2Byte(&byteBuffer)
					posY := utils.ReadAs2Byte(&byteBuffer)
					player.SetLocation(posX, posY)
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}

func HttpListener(w http.ResponseWriter, r *http.Request) {
	var jsonMap map[string]any
	err := json.NewDecoder(r.Body).Decode(&jsonMap)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", jsonMap["alert"])

	fmt.Println(jsonMap["alert"].(string))
}

func loadTextures() {
	img, _ := g.LoadImage("static/img/map.png")
	g.NewTextureFromRgba(img, func(tex *g.Texture) {
		texture = tex
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	go clientConnectionListener(":19003")

	http.HandleFunc("/", HttpListener)
	go http.ListenAndServe(":19006", nil)

	for i := 0; i < 169; i++ {
		addPlayerByName(fmt.Sprintf("bot_%02x", rand.Int31()), true)
	}

	wnd := g.NewMasterWindow("PWNBG Game Server", 900, 600, 0)

	loadTextures()

	wnd.Run(loop)
}
