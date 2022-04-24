package game

type ItemType string

const (
	HAT   ItemType = "hat"
	SHIRT ItemType = "shirt"
	PANTS ItemType = "pants"
	GUN   ItemType = "gun"
)

type Item struct {
	Id      int      `json:"id"`
	Name    string   `json:"name"`
	Type    ItemType `json:"type"`
	Level   int      `json:"level"`
	StatMin int      `json:"statMin"`
	StatMax int      `json:"statMax"`
}
