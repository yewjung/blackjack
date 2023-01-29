package main

import (
	"fmt"
	"strconv"
	"sync"
)

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
	Players []*Player
	Turn    int
}

func (game *Game) DrawCard() {
	currentPlayer := game.Players[game.Turn]
	gameData := &currentPlayer.GameData

	card := game.Deck[len(game.Deck)-1]
	game.Deck = game.Deck[:len(game.Deck)-1]

	gameData.Hand = append(gameData.Hand, card)

	game.broadcastCardOtherPlayers(card)
}

func broadcastCard(currentPlayer *Player, card Card, wg *sync.WaitGroup) {
	// TODO: implement broadcardCard
	wg.Done()
}

func (game *Game) broadcastCardOtherPlayers(card Card) {
	wg := sync.WaitGroup{}
	wg.Add(len(game.Players))
	// broadcast to current player about his new card
	currentPlayer := game.Players[game.Turn]
	go broadcastCard(currentPlayer, card, &wg)

	// broadcast to other players about current players new card
	// (but most not tell them the card details)
	for i := 0; i < len(game.Players); i++ {
		if game.Turn == i {
			continue
		}
		player := game.Players[i]
		go func(player *Player) {
			// TODO: send a proper response here
			response := fmt.Sprintf("new card for player[%s]", strconv.Itoa(game.Turn))
			player.Conn.WriteJSON(response)
			wg.Done()
		}(player)
	}
	wg.Wait()

}
func (game *Game) Stand() {
	game.broadcastStand()
	game.Turn += 1
}

func (game *Game) broadcastStand() {
	wg := sync.WaitGroup{}
	wg.Add(len(game.Players))

	// broadcast to current player about his stand
	currentPlayer := game.Players[game.Turn]
	go broadcastStand(currentPlayer, &wg)

	// broadcast to other players about current players stand
	for i := 0; i < len(game.Players); i++ {
		if game.Turn == i {
			continue
		}
		player := game.Players[i]
		go func(player *Player) {
			// TODO: send a proper response here
			response := fmt.Sprintf("player[%s] stands", strconv.Itoa(game.Turn))
			player.Conn.WriteJSON(response)
			wg.Done()
		}(player)
	}
	wg.Wait()
}

func broadcastStand(currentPlayer *Player, wg *sync.WaitGroup) {
	// TODO: implement this method
	wg.Done()
}

func (game *Game) StartGame() {
	// prepare deck by shuffling

	// send two cards to each player

}
