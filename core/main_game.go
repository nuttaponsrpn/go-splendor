package core

import (
	"errors"
	"log"
	"math/rand"
	"slices"
	"strconv"

	"github.com/nuttaponsrpn/go-splendor/gotype"
)

type WebsocketPlayerAction struct {
	PlayerId      string                 `json:"playerId"`
	SelectedGems  []gotype.GemType       `json:"selectedGems"`
	PurchasedCard gotype.DevelopmentCard `json:"purchasedCard"`
	ReservedCard  gotype.DevelopmentCard `json:"reservedCard"`
	Status        gotype.Status          `json:"status"`
}

type GameService interface {
	GetGameState() gotype.GameState
	JoinPlayer(playerId string)
	UpdateGameState(action WebsocketPlayerAction) error
}

type GameServiceImpl struct {
	GameState gotype.GameState
}

func NewGameService(GameState gotype.GameState) GameService {
	return &GameServiceImpl{GameState: GameState}
}

func (s *GameServiceImpl) GetGameState() gotype.GameState {
	fmtGameState := s.GameState
	tiles := fmtGameState.DevelopmentTiles

	if len(tiles.Level1) > 0 {
		fmtGameState.DevelopmentTiles.Level1 = tiles.Level1[0:4]
		fmtGameState.DevelopmentTiles.Level2 = tiles.Level2[0:4]
		fmtGameState.DevelopmentTiles.Level3 = tiles.Level3[0:4]
	}

	return fmtGameState
}

func (s *GameServiceImpl) JoinPlayer(playerId string) {
	players := s.GameState.Players
	isPlayerExist := slices.ContainsFunc(players, func(p gotype.Player) bool {
		return p.Id == playerId
	})

	if len(s.GameState.Players) == 0 {
		s.GameState.CurrentPlayerId = playerId
	}

	if isPlayerExist {
		return
	}

	if s.GameState.DevelopmentTiles.Level1 == nil {
		InitGameCard(&s.GameState)
	}

	newPlayer := gotype.Player{
		Id: playerId,
		Gems: map[gotype.GemType]int{
			gotype.Diamond:  0,
			gotype.Sapphire: 0,
			gotype.Emerald:  0,
			gotype.Ruby:     0,
			gotype.Onyx:     0,
			gotype.Joker:    0,
		},
		Points:        0,
		ReservedCards: []gotype.DevelopmentCard{},
		PurchaseCards: []gotype.DevelopmentCard{},
		NobleCards:    []gotype.NobleCard{},
	}
	s.GameState.Players = append(s.GameState.Players, newPlayer)
}

func InitGameCard(game *gotype.GameState) {
	developmentTiles, nobles := RandomCards()
	game.Nobles = nobles
	game.DevelopmentTiles = *developmentTiles
	game.Gems = map[gotype.GemType]int{
		gotype.Diamond:  7,
		gotype.Sapphire: 7,
		gotype.Emerald:  7,
		gotype.Ruby:     7,
		gotype.Onyx:     7,
		gotype.Joker:    5,
	}
}

func RandomCards() (*gotype.DevelopmentTiles, []gotype.NobleCard) {
	developmentTiles := &gotype.DevelopmentTiles{
		Level1: DevelopmentLevel1,
		Level2: DevelopmentLevel2,
		Level3: DevelopmentLevel3,
	}

	ShuffleCard(developmentTiles.Level1)
	ShuffleCard(developmentTiles.Level2)
	ShuffleCard(developmentTiles.Level3)

	nobles := Nobles

	ShuffleCard(nobles)
	nobles = nobles[0:4]

	return developmentTiles, nobles
}

func ShuffleCard[T gotype.DevelopmentCard | gotype.NobleCard](card []T) {
	for i := range card {
		j := rand.Intn(i + 1)
		card[i], card[j] = card[j], card[i]
	}
}

func GetMessage() {}

func (s *GameServiceImpl) UpdateGameState(Action WebsocketPlayerAction) error {
	playerIndex := slices.IndexFunc(s.GameState.Players, func(p gotype.Player) bool { return p.Id == Action.PlayerId })

	if playerIndex == -1 {
		return errors.New("not found player: " + Action.PlayerId)
	}

	currentPlayer := &s.GameState.Players[playerIndex]

	s.UpdatePlayerGems(currentPlayer, Action.SelectedGems)
	s.UpdatedPlayerPurchasedCard(currentPlayer, Action.PurchasedCard)
	s.UpdatedPlayerReservedCard(currentPlayer, Action.ReservedCard)
	s.AddNobleCard(currentPlayer)
	currentPlayer.Points = CalculatePoints(currentPlayer.PurchaseCards, currentPlayer.NobleCards)

	if err := s.UpdateNextPlayer(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *GameServiceImpl) UpdatePlayerGems(currentPlayer *gotype.Player, selectedGems []gotype.GemType) {
	for _, gems := range selectedGems {
		currentPlayer.Gems[gems] += 1
		s.GameState.Gems[gems] -= 1
	}
}

func (s *GameServiceImpl) UpdatedPlayerPurchasedCard(currentPlayer *gotype.Player, card gotype.DevelopmentCard) {
	cardCost := CalculatePayCostReducePurchaseCard(*currentPlayer, card)

	for gemType := range currentPlayer.Gems {
		if cardCost[gemType] > 0 {
			if currentPlayer.Gems[gemType] >= cardCost[gemType] {
				// Have token enough for paying then return gems to game state
				s.GameState.Gems[gemType] += cardCost[gemType]
				currentPlayer.Gems[gemType] -= cardCost[gemType]
			} else {
				// Not enough token pay existing gems + joker gems
				payWithJoker := currentPlayer.Gems[gemType] - cardCost[gemType]
				s.GameState.Gems[gemType] += currentPlayer.Gems[gemType]
				currentPlayer.Gems[gemType] -= currentPlayer.Gems[gemType]
				currentPlayer.Gems[gotype.Joker] += payWithJoker
			}
		}
	}
	isReservedCard := slices.IndexFunc(currentPlayer.ReservedCards, func(rCard gotype.DevelopmentCard) bool {
		return rCard.ID == card.ID
	})

	if isReservedCard != -1 {
		if fCard, err := FilterCard(currentPlayer.ReservedCards, card.ID); err == nil {
			currentPlayer.ReservedCards = fCard
			currentPlayer.PurchaseCards = append(currentPlayer.PurchaseCards, card)
		}
	} else {
		switch card.Level {
		case 1:
			if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level1, card.ID); err == nil {
				s.GameState.DevelopmentTiles.Level1 = fCard
				currentPlayer.PurchaseCards = append(currentPlayer.PurchaseCards, card)
			} else {
				log.Fatal(err)
			}
		case 2:
			if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level2, card.ID); err == nil {
				s.GameState.DevelopmentTiles.Level2 = fCard
				currentPlayer.PurchaseCards = append(currentPlayer.PurchaseCards, card)
			} else {
				log.Fatal(err)
			}
		case 3:
			if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level3, card.ID); err == nil {
				s.GameState.DevelopmentTiles.Level3 = fCard
				currentPlayer.PurchaseCards = append(currentPlayer.PurchaseCards, card)
			} else {
				log.Fatal(err)
			}
		}
	}
}

func CalculatePayCostReducePurchaseCard(currentPlayer gotype.Player, card gotype.DevelopmentCard) map[gotype.GemType]int {
	var DiamondCost = card.Cost.Diamond - CalculateCardGems(currentPlayer.PurchaseCards, gotype.Diamond)
	var SapphireCost = card.Cost.Sapphire - CalculateCardGems(currentPlayer.PurchaseCards, gotype.Sapphire)
	var EmeraldCost = card.Cost.Emerald - CalculateCardGems(currentPlayer.PurchaseCards, gotype.Emerald)
	var RubyCost = card.Cost.Ruby - CalculateCardGems(currentPlayer.PurchaseCards, gotype.Ruby)
	var OnyxCost = card.Cost.Onyx - CalculateCardGems(currentPlayer.PurchaseCards, gotype.Onyx)
	if DiamondCost <= 0 {
		DiamondCost = 0
	}
	if SapphireCost <= 0 {
		SapphireCost = 0
	}
	if EmeraldCost <= 0 {
		EmeraldCost = 0
	}
	if RubyCost <= 0 {
		RubyCost = 0
	}
	if OnyxCost <= 0 {
		OnyxCost = 0
	}

	return map[gotype.GemType]int{
		gotype.Diamond:  DiamondCost,
		gotype.Sapphire: SapphireCost,
		gotype.Emerald:  EmeraldCost,
		gotype.Ruby:     RubyCost,
		gotype.Onyx:     OnyxCost,
	}
}

func (s *GameServiceImpl) UpdatedPlayerReservedCard(currentPlayer *gotype.Player, card gotype.DevelopmentCard) {
	switch card.Level {
	case 1:
		if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level1, card.ID); err == nil {
			s.GameState.DevelopmentTiles.Level1 = fCard
			currentPlayer.ReservedCards = append(currentPlayer.ReservedCards, card)
		} else {
			log.Fatal(err)
		}
	case 2:
		if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level2, card.ID); err == nil {
			s.GameState.DevelopmentTiles.Level2 = fCard
			currentPlayer.ReservedCards = append(currentPlayer.ReservedCards, card)
		} else {
			log.Fatal(err)
		}
	case 3:
		if fCard, err := FilterCard(s.GameState.DevelopmentTiles.Level3, card.ID); err == nil {
			s.GameState.DevelopmentTiles.Level3 = fCard
			currentPlayer.ReservedCards = append(currentPlayer.ReservedCards, card)
		} else {
			log.Fatal(err)
		}
	}
}

func (s *GameServiceImpl) AddNobleCard(currentPlayer *gotype.Player) {
	var removeNobleIndex = -1
	for index, noble := range s.GameState.Nobles {
		diamondPass := CalculateCardGems(currentPlayer.PurchaseCards, gotype.Diamond) >= noble.Cost[gotype.Diamond]
		saphirePass := CalculateCardGems(currentPlayer.PurchaseCards, gotype.Sapphire) >= noble.Cost[gotype.Sapphire]
		emeraldPass := CalculateCardGems(currentPlayer.PurchaseCards, gotype.Emerald) >= noble.Cost[gotype.Emerald]
		rubyPass := CalculateCardGems(currentPlayer.PurchaseCards, gotype.Ruby) >= noble.Cost[gotype.Ruby]
		onyxPass := CalculateCardGems(currentPlayer.PurchaseCards, gotype.Onyx) >= noble.Cost[gotype.Onyx]

		if diamondPass && saphirePass && emeraldPass && rubyPass && onyxPass {
			currentPlayer.NobleCards = append(currentPlayer.NobleCards, noble)
			removeNobleIndex = index
			currentPlayer.Points = currentPlayer.Points + noble.Points
			break
		}
	}

	if removeNobleIndex != -1 {
		s.GameState.Nobles = append(s.GameState.Nobles[:removeNobleIndex], s.GameState.Nobles[removeNobleIndex+1:]...)
	}
}

func FilterCard(card []gotype.DevelopmentCard, removeId int) ([]gotype.DevelopmentCard, error) {
	removeIndex := slices.IndexFunc(card, func(c gotype.DevelopmentCard) bool { return c.ID == removeId })

	if removeIndex == -1 {
		return nil, errors.New("filter card not found removeId" + strconv.Itoa(int(removeId)))
	}
	return append(card[:removeIndex], card[removeIndex+1:]...), nil
}

func CalculatePoints(purchasedCards []gotype.DevelopmentCard, noble []gotype.NobleCard) int {
	points := int(0)
	for _, card := range purchasedCards {
		points = points + card.Points
	}
	for _, card := range noble {
		points = points + card.Points
	}
	return points
}

func CalculateCardGems(purchasedCards []gotype.DevelopmentCard, gemType gotype.GemType) int {
	gemsPoints := int(0)
	for _, card := range purchasedCards {
		if card.GemType == gemType {
			gemsPoints = gemsPoints + 1
		}
	}
	return gemsPoints
}

func (s *GameServiceImpl) UpdateNextPlayer() error {
	if s.GameState.CurrentPlayerId == "" {
		s.GameState.CurrentPlayerId = s.GameState.Players[0].Id
		return nil
	}

	currentIndex := slices.IndexFunc(s.GameState.Players, func(player gotype.Player) bool {
		return player.Id == s.GameState.CurrentPlayerId
	})

	if currentIndex == -1 {
		return errors.New("not found next player(current player): " + s.GameState.CurrentPlayerId)
	}

	nextIndex := currentIndex + 1
	lastIndex := len(s.GameState.Players) - 1
	if nextIndex > lastIndex {
		nextIndex = 0
	}
	s.GameState.CurrentPlayerId = s.GameState.Players[nextIndex].Id
	return nil
}

func CalcualteWinner() {}
