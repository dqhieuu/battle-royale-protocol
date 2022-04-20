package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"fps-game-server/src/game"
	g "github.com/AllenDang/giu"
	"github.com/xtaci/kcp-go"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

var texture *g.Texture

var selectedRoomIndex int
var playerRoomLookupByName map[string]*game.GameRoom
var rooms []*game.GameRoom

func addPlayerByName(name string) *game.Player {
	if playerRoomLookupByName[name] != nil {
		return playerRoomLookupByName[name].PlayerByName(name)
	}

	hasEmptyRoom := false
	for i := range rooms {
		if rooms[i].IsFull() {
			continue
		}

		hasEmptyRoom = true
		newPlayer := game.NewPlayerWithName(name)
		newPlayer.SetMap(rooms[i].Map)
		rooms[i].AddPlayer(newPlayer)
		playerRoomLookupByName[name] = rooms[i]
		return newPlayer
	}

	if !hasEmptyRoom {
		newRoom := game.NewGameRoom()
		go newRoom.Activate()

		newPlayer := game.NewPlayerWithName(name)
		newPlayer.SetMap(newRoom.Map)
		newRoom.AddPlayer(newPlayer)
		rooms = append(rooms, newRoom)
		playerRoomLookupByName[name] = newRoom
		return newPlayer
	}
	return nil
}

func loop() {
	var lobbyTableRows []*g.TableRowWidget

	var roomTableRows []*g.TableRowWidget

	for i := range rooms {
		i := i
		lobbyTableRows = append(lobbyTableRows, g.TableRow(
			g.Selectable(strconv.Itoa(i)).Flags(g.SelectableFlagsSpanAllColumns).OnClick(func() {
				println(i)
				selectedRoomIndex = i
			}),
			g.Label(fmt.Sprintf("%d/%d", rooms[i].PlayerCount(), rooms[i].MaxPlayers)),
			g.Label(strconv.Itoa(int(rooms[i].State))),
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

		roomTableRows = append(roomTableRows,
			g.TableRow(
				g.Selectable(strconv.Itoa(i+1)).Flags(g.SelectableFlagsSpanAllColumns),
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
		g.Table().Columns(
			g.TableColumn("Phòng"),
			g.TableColumn("Người chơi"),
			g.TableColumn("Trạng thái"),
		).Rows(
			lobbyTableRows...,
		).Freeze(1, 1).Size(300, 100),
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
		).Size(600, 200),

		g.Custom(func() {
			canvas := g.GetCanvas()
			pos := g.GetCursorScreenPos()
			red := color.RGBA{200, 75, 75, 255}
			blue := color.RGBA{75, 75, 200, 255}
			if texture != nil {
				canvas.AddImage(texture, pos, pos.Add(image.Point{400, 400}))
			}
			for i := range players {
				circlePos := pos.Add(image.Point{int(players[i].Location.X), int(players[i].Location.Y)})
				canvas.AddCircleFilled(circlePos, 5, red)
				canvas.AddCircleFilled(circlePos, 3, color.White)
				if players[i].IsDead() {
					canvas.AddCircleFilled(circlePos, 3, color.Black)
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
								pos.Add(image.Point{int(players[i].Location.X), int(players[i].Location.Y)}), pos.Add(image.Point{int(players[j].Location.X), int(players[j].Location.Y)}), blue, 3)
							break
						}
					}
				}
			}

			g.Update()
		}),
	)
}

func server(url string) {
	if listener, err := kcp.Listen(url); err == nil {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleConn(s)
		}
	} else {
		log.Fatal(err)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		println(conn)
		buf := []byte("test")
		if _, err := conn.Write(buf[:]); err != nil {
			log.Println(err)
			return
		}

		time.Sleep(time.Millisecond * 5)
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

func main() {
	rand.Seed(time.Now().UnixNano())
	playerRoomLookupByName = make(map[string]*game.GameRoom)

	go server("127.0.0.1:19001")

	http.HandleFunc("/", HttpListener)
	go http.ListenAndServe("127.0.0.1:19006", nil)

	for i := 0; i < 150; i++ {
		addPlayerByName(fmt.Sprintf("%02x", rand.Int31()))
	}

	wnd := g.NewMasterWindow("Canvas", 600, 600, 0)

	img, _ := g.LoadImage("static/img/map.png")
	g.NewTextureFromRgba(img, func(tex *g.Texture) {
		texture = tex
	})

	wnd.Run(loop)
}
