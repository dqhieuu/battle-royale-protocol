package game

import (
	"math/rand"
	"sync"
	"time"
)

const StateTransitionDelay = 5 * time.Second

const MaxShootableDistance = 40
const SqrMaxShootableDistance = MaxShootableDistance * MaxShootableDistance

var MaxPlayers = 100
var MinTriggerPlayingPlayers = 25

type RoomState int

const (
	WAITING RoomState = iota
	PLAYING
	FINISHED
)

type GameRoom struct {
	Id                 int
	MaxPlayers         int
	Map                *GameMap
	Players            []*Player
	playerLookupByName map[string]*Player
	State              RoomState
	mutex              *sync.RWMutex
}

func NewGameRoom() *GameRoom {
	return &GameRoom{
		Id:                 rand.Int(),
		MaxPlayers:         MaxPlayers,
		Map:                NewGameMap(DefaultMapWidth, DefaultMapHeight),
		Players:            []*Player{},
		playerLookupByName: map[string]*Player{},
		State:              WAITING,
		mutex:              &sync.RWMutex{},
	}
}

func (room *GameRoom) AddPlayer(player *Player) bool {
	room.mutex.Lock()
	if len(room.Players) >= room.MaxPlayers || room.State != WAITING {
		return false
	}

	room.Players = append(room.Players, player)
	room.playerLookupByName[player.Name] = player
	room.mutex.Unlock()

	return true
}

func (room *GameRoom) PlayerCount() int {
	return len(room.Players)
}

func (room *GameRoom) PlayerByName(name string) *Player {
	return room.playerLookupByName[name]
}

func (room *GameRoom) IsFull() bool {
	return len(room.Players) >= room.MaxPlayers
}

func (room *GameRoom) Activate() {
	for room.State != FINISHED {
		room.mutex.Lock()
		if room.State == WAITING {
			if room.PlayerCount() >= MinTriggerPlayingPlayers {
				time.Sleep(StateTransitionDelay)
				room.State = PLAYING
			}
		} else if room.State == PLAYING {
			aliveCount := 0

			for i := range room.Players {
				pi := room.Players[i]
				if pi.IsDead() {
					continue
				} else {
					aliveCount++
				}

				pi.UpdateLocation()

				if rand.Intn(200) < 1 { // 0.5% chance to get a new equipment
					pi.Inventory.EquipItem(room.Map.ItemStore.GetRandomItem())
				}

				for j := range room.Players {
					pj := room.Players[j]
					if i == j || pj.IsDead() {
						continue
					}
					xd := pi.Location.X - pj.Location.X
					yd := pi.Location.Y - pj.Location.Y
					sqrDistance := xd*xd + yd*yd

					if sqrDistance < SqrMaxShootableDistance {
						if rand.Intn(100) < 3 { // 3% chance to shoot
							damage := 1
							weaponDmg := pi.GetDamage() - pj.GetDefense()
							if weaponDmg > damage {
								damage = weaponDmg
							}

							pj.DecreaseHP(damage, pi)
						}
						break
					}
				}
			}

			if aliveCount <= 1 {
				time.Sleep(StateTransitionDelay)
				room.State = FINISHED
			}
		}
		room.mutex.Unlock()
		time.Sleep(time.Millisecond * 15)
	}
}
