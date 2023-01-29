package main

type Suit int

const (
	HEARTS Suit = iota
	SPADES
	CLUBS
	DIAMONDS
)

type Card struct {
	Rank int
	Suit Suit
}

type GameData struct {
	Hand []Card
}

func (data *GameData) value() int {
	return 0
}

type Game struct {
	Deck    []Card
	Players []Player
	Turn    int
}

func (game *Game) DrawCard() {}
func (game *Game) Stand()    {}
