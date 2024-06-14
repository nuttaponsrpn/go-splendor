package gotype

type GameState struct {
	CurrentPlayerId  string           `json:"currentPlayerId"`
	Players          []Player         `json:"players"`
	Gems             map[GemType]int  `json:"gems"`
	Nobles           []NobleCard      `json:"nobles"`
	DevelopmentTiles DevelopmentTiles `json:"developmentTiles"`
	State            Status           `json:"state"`
}

type Player struct {
	Id            string            `json:"id"`
	Gems          map[GemType]int   `json:"gems"`
	Points        int               `json:"points"`
	ReservedCards []DevelopmentCard `json:"reservedCards"`
	PurchaseCards []DevelopmentCard `json:"purchasedCards"`
	NobleCards    []NobleCard       `json:"nobleCards"`
}

type Gems struct {
	Diamond  int `json:"diamond"`
	Sapphire int `json:"sapphire"`
	Emerald  int `json:"emerald"`
	Ruby     int `json:"ruby"`
	Onyx     int `json:"onyx"`
	Joker    int `json:"joker"`
}

type GemType string

const (
	Diamond  GemType = "diamond"
	Sapphire GemType = "sapphire"
	Emerald  GemType = "emerald"
	Ruby     GemType = "ruby"
	Onyx     GemType = "onyx"
	Joker    GemType = "joker"
)

type DevelopmentTiles struct {
	Level1 []DevelopmentCard `json:"level1"`
	Level2 []DevelopmentCard `json:"level2"`
	Level3 []DevelopmentCard `json:"level3"`
}

type DevelopmentCard struct {
	ID      int     `json:"id"`
	Level   int     `json:"level"`
	Cost    Gems    `json:"cost"`
	Points  int     `json:"points"`
	GemType GemType `json:"gemType"`
}

type NobleCard struct {
	ID     int             `json:"id"`
	Cost   map[GemType]int `json:"cost"`
	Points int             `json:"points"`
}

type Status string

const (
	Waiting         Status = "Waiting"
	Started         Status = "Started"
	End             Status = "End"
	CloseConnection Status = "CloseConnection"
)
