package game

import (
	"fmt"
	"fps-client/src/utils"
	"github.com/golang/geo/r2"
	"math"
	"math/rand"
)

func (i Item) String() string {
	return fmt.Sprintf("ID:%d Name:%s Type:%s Level:%d, StatMin:%d Statmax:%d \n", i.Id, i.Name, i.Type, i.Level, i.StatMin, i.StatMax)
}

type Inventory struct {
	Hat   *Item
	Shirt *Item
	Pants *Item
	Gun   *Item
}

// EquipItem returns bool if the current item has changed
func (i *Inventory) EquipItem(item *Item) bool {
	if item == nil {
		return false
	}

	var currentItemPtr **Item
	switch item.Type {
	case HAT:
		currentItemPtr = &i.Hat
	case SHIRT:
		currentItemPtr = &i.Shirt
	case PANTS:
		currentItemPtr = &i.Pants
	case GUN:
		currentItemPtr = &i.Gun
	}

	if *currentItemPtr == nil || (*currentItemPtr).Level < item.Level {
		*currentItemPtr = item
		return true
	}
	return false
}

type Player struct {
	sameDirectionCounter int
	gameMap              *GameMap
	HP                   int
	Name                 string
	Inventory            Inventory
	Location             r2.Point
	Direction            r2.Point // 2D normalized vector
	IsBot                bool
}

func NewPlayer() *Player {
	return &Player{HP: 100, Name: fmt.Sprintf("%02x", rand.Int31())}
}

func NewBotPlayerWithName(name string) *Player {
	return &Player{HP: 100, Name: name, IsBot: true}
}

func NewPlayerWithName(name string) *Player {
	return &Player{HP: 100, Name: name}
}

func (p *Player) chooseNewRandomDirection() {
	nextDeg := rand.Float64() * math.Pi * 2
	nextDir := r2.Point{X: math.Cos(nextDeg), Y: math.Sin(nextDeg)}
	p.Direction = nextDir
}

func (p *Player) SetMap(gameMap *GameMap) {
	p.gameMap = gameMap
	p.Location = r2.Point{float64(rand.Intn(gameMap.Width)), float64(rand.Intn(gameMap.Height))}
}

func (p *Player) SetLocation(x, y int) {
	if p.IsDead() {
		return
	}
	p.Location = r2.Point{X: float64(x), Y: float64(y)}
}

func (p *Player) UpdateLocation() {
	if p.IsDead() || !p.IsBot {
		return
	}

	p.sameDirectionCounter--
	if p.sameDirectionCounter <= 0 {
		p.chooseNewRandomDirection()
		p.sameDirectionCounter = utils.RandRangeInt(30, 60)
	}

	const MinStep = 0.3
	const MaxStep = 0.5
	nextLocation := p.Location.Add(p.Direction.Mul(utils.RandRangeFloat(MinStep, MaxStep)))
	if p.gameMap.BBox.ContainsPoint(nextLocation) {
		p.Location = nextLocation
	} else {
		p.chooseNewRandomDirection()
	}
}

func (p *Player) DecreaseHP(amount int, fromPlayer *Player) {
	if p.IsDead() {
		return
	}

	p.HP -= amount
	if p.IsDead() {
		p.HP = 0
		//fmt.Printf("Player %s is killed by %s\n", p.Name, fromPlayer.Name)
	}
}

func (p *Player) IsDead() bool {
	return p.HP <= 0
}

func (p *Player) GetDamage() int {
	damageMin := 0
	damageMax := 0
	if p.Inventory.Gun != nil {
		damageMin = p.Inventory.Gun.StatMin
		damageMax = p.Inventory.Gun.StatMax
	}
	return 3 + utils.RandRangeInt(damageMin, damageMax) // 3 is base damage
}

func (p *Player) GetDefense() int {
	armorMin := 0
	armorMax := 0
	if p.Inventory.Hat != nil {
		armorMin += p.Inventory.Hat.StatMin
		armorMax += p.Inventory.Hat.StatMax
	}
	if p.Inventory.Shirt != nil {
		armorMin += p.Inventory.Shirt.StatMin
		armorMax += p.Inventory.Shirt.StatMax
	}
	if p.Inventory.Pants != nil {
		armorMin += p.Inventory.Pants.StatMin
		armorMax += p.Inventory.Pants.StatMax
	}
	return utils.RandRangeInt(armorMin, armorMax)
}
