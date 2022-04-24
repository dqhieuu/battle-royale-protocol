package game

import (
	"encoding/json"
	"fmt"
	"github.com/golang/geo/r2"
	"io"
	"math/rand"
	"os"
	"sync"
)

const (
	DefaultMapWidth  = 400
	DefaultMapHeight = 400
)

type ItemStore struct {
	store     sync.Map
	allKeys   []int
	gunKeys   []int
	hatKeys   []int
	shirtKeys []int
	pantsKeys []int
}

var Items *ItemStore

func (is *ItemStore) QueryItem(id int) *Item {
	item, ok := is.store.Load(id)

	if !ok {
		return nil
	}

	return item.(*Item)
}

func (is *ItemStore) GetRandomItem() *Item {
	itemCount := len(is.allKeys)
	if itemCount <= 0 {
		return nil
	}

	randKey := rand.Intn(itemCount)
	item, _ := is.store.Load(is.allKeys[randKey])
	return item.(*Item)
}

func (is *ItemStore) loadItem(path string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(jsonFile)

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("failed to read json file, error: %v", err)
		return
	}

	var data []Item
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Printf("failed to unmarshal json file, error: %v", err)
		return
	}

	// Print
	for _, item := range data {
		is.store.Store(item.Id, &item) // add to map
		is.allKeys = append(is.allKeys, item.Id)
		switch item.Type {
		case HAT:
			is.hatKeys = append(is.hatKeys, item.Id)
		case PANTS:
			is.pantsKeys = append(is.pantsKeys, item.Id)
		case SHIRT:
			is.shirtKeys = append(is.shirtKeys, item.Id)
		case GUN:
			is.gunKeys = append(is.gunKeys, item.Id)
		}
	}
}

func NewItemStore() bool {
	Items = &ItemStore{}
	Items.loadItem("static/items.json")
	return true
}

type GameMap struct {
	Width     int
	Height    int
	BBox      r2.Rect
	ItemStore *ItemStore
}

func NewGameMap(width, height int) *GameMap {
	if Items == nil {
		NewItemStore()
	}
	var gameMap = GameMap{Width: width, Height: height, BBox: r2.RectFromPoints(r2.Point{0, 0}, r2.Point{float64(width), float64(height)}), ItemStore: Items}
	return &gameMap
}
