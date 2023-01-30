package main

import (
	"sync"
)

type Suit int

const (
	HEARTS Suit = iota
	SPADES
	CLUBS
	DIAMONDS
)

type GameEvent int

const (
	HIT GameEvent = iota
	STAND
	NEXT
)

type GameResponse struct {
	Event        GameEvent `json:"event"`
	AffectedUser string    `json:"affectedUser"`
	NewCard      *Card     `json:"card,omitempty"`
}

type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

type Rank struct {
	Value int    `json:"-"`
	Name  string `json:"name"`
}

type Result int

const (
	WIN Result = iota
	LOSE
	DRAW
)

type GameData struct {
	Hand []Card
}

func getHandValue(hand []Card) int {
	value := 0
	for _, card := range hand {
		if card.Rank.Value == 1 && len(hand) <= 2 {
			value += 11
		} else {
			value += card.Rank.Value
		}
	}
	return value
}

type Game struct {
	Deck        []Card
	Players     []*Player
	Turn        int
	DealerHands []Card
}

func (game *Game) DrawCard(player *Player) {
	if !game.isPlayersTurn(player) {
		// it is not his turn
		return
	}
	currentPlayer := game.Players[game.Turn]
	gameData := &currentPlayer.GameData

	card := game.Deck[len(game.Deck)-1]
	game.Deck = game.Deck[:len(game.Deck)-1]

	gameData.Hand = append(gameData.Hand, card)

	game.broadcast(HIT, &card)
}

func (game *Game) isPlayersTurn(player *Player) bool {
	return player.ID == game.Players[game.Turn].ID
}

func (game *Game) Stand(player *Player) {
	if !game.isPlayersTurn(player) {
		// it is not his turn
		return
	}
	game.broadcast(STAND, nil)
	game.nextTurn()
}

func (game *Game) nextTurn() {
	game.Turn += 1
	if game.Turn == len(game.Players) {
		// dealer's turn to draw
		// broadcast to players that it is dealer's turn
		return
	}

	// broadcast to players about next player's turn

}

func (game *Game) broadcast(event GameEvent, newCard *Card) {
	wg := sync.WaitGroup{}
	wg.Add(len(game.Players))

	// broadcast to current player
	currentPlayer := game.Players[game.Turn]
	go func() {
		response := getEventResponse(event, true, currentPlayer.ID, newCard)
		currentPlayer.Conn.WriteJSON(response)
		wg.Done()
	}()
	for i := 0; i < len(game.Players); i++ {
		if game.Turn == i {
			continue
		}
		player := game.Players[i]
		go func(player *Player) {
			response := getEventResponse(event, false, currentPlayer.ID, newCard)
			player.Conn.WriteJSON(response)
			wg.Done()
		}(player)
	}
	wg.Wait()
}

func getEventResponse(event GameEvent, isCurrentPlayer bool, affectedUserId string, newCard *Card) GameResponse {
	card := newCard
	if !isCurrentPlayer {
		card = nil
	}
	return GameResponse{
		Event:        event,
		AffectedUser: affectedUserId,
		NewCard:      card,
	}
}

func (game *Game) calculateHands() {
	dealerValue := getHandValue(game.DealerHands)
	wg := sync.WaitGroup{}
	wg.Add(len(game.Players))
	for _, player := range game.Players {
		go func(player *Player) {
			broadcastResults(player, dealerValue)
			wg.Done()
		}(player)
	}
	wg.Wait()
}

func broadcastResults(player *Player, dealerValue int) {
	playerValue := getHandValue(player.GameData.Hand)
	if playerValue > 21 {
		// "Player busts, dealer wins"
		broadcastResultToPlayer(player, LOSE)
	} else if dealerValue > 21 {
		// "Dealer busts, player wins"
		broadcastResultToPlayer(player, WIN)
	} else if dealerValue > playerValue {
		// "Dealer wins"
		broadcastResultToPlayer(player, LOSE)
	} else if playerValue > dealerValue {
		// "Player wins"
		broadcastResultToPlayer(player, WIN)
	} else {
		// "It's a tie"
		broadcastResultToPlayer(player, DRAW)
	}
}

func broadcastResultToPlayer(player *Player, result Result) {
	switch result {
	case WIN:
		// broadcast win
	case LOSE:
		// broadcast lost
	case DRAW:
		// broadcast draw
	}
}

func (game *Game) StartGame() {
	// prepare deck by shuffling

	// send two cards to each player

	// send two cards to dealer

	// ==== LOOP THROUGH PLAYERS ====
	// player either hit or stand
	// player allowed to have max of 5 cards
	// if stand or 5 cards reached, move to next player

	// ==== LOOP ENDS ====

	// dealer either hit or stand

	// once dealer is done, calculate values for each player
	// compare hand value of each player to dealer

	// GAME ENDS

}

// IDEA:
// when player sends message, check if it his turn
// if it is not his turn, ignore his message
// only proceed the game when the right player sends a message
