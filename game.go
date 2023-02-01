package main

import (
	"math/rand"
	"sync"
	"time"
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
	NextPlayer   string    `json:"nextPlayer"`
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

	game.broadcast(PLAYER_HIT, &card)
}

func (game *Game) isPlayersTurn(player *Player) bool {
	return player.ID == game.Players[game.Turn].ID
}

func (game *Game) Stand(player *Player) {
	if !game.isPlayersTurn(player) {
		// it is not his turn
		return
	}
	game.broadcast(PLAYER_STAND, nil)
	game.nextTurn()
}

func (game *Game) nextTurn() {
	game.Turn += 1
	if game.Turn == len(game.Players) {
		// dealer's turn to draw
		// broadcast to players that it is dealer's turn
		game.endGame()
		return
	}

	// broadcast to players about next player's turn
	game.broadcastNextTurn()
}

func (game *Game) broadcastNextTurn() {
	nextPlayer := game.Players[game.Turn]
	response := Response{
		Event: YOUR_TURN,
	}
	wg := sync.WaitGroup{}
	wg.Add(len(game.Players))
	go func() {
		nextPlayer.Conn.WriteJSON(response)
		wg.Done()
	}()
	for _, player := range game.Players {
		if player.ID == nextPlayer.ID {
			continue
		}
		go func(player *Player) {
			response := Response{
				Event:      NEXT_PLAYER,
				NextPlayer: nextPlayer.ID,
			}
			player.Conn.WriteJSON(response)
			wg.Done()
		}(player)
	}
}

func (game *Game) endGame() {
	game.calculateHands()
}

func (game *Game) broadcast(event Event, newCard *Card) {
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

func getEventResponse(event Event, isCurrentPlayer bool, affectedUserId string, newCard *Card) Response {
	card := newCard
	if !isCurrentPlayer {
		card = nil
	}
	return Response{
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
		response := Response{
			Event: GAME_WIN,
		}
		player.Conn.WriteJSON(response)
	case LOSE:
		// broadcast lost
		response := Response{
			Event: GAME_LOST,
		}
		player.Conn.WriteJSON(response)
	case DRAW:
		// broadcast draw
		response := Response{
			Event: GAME_DRAW,
		}
		player.Conn.WriteJSON(response)
	}
}

func (game *Game) StartGame() {
	// prepare deck by shuffling
	game.Deck = shuffle(game.Deck)
	// send two cards to each player

	// send two cards to dealer

}

func shuffle[T any](list []T) []T {
	rand.Seed(time.Now().UnixNano())
	for i := len(list) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		list[i], list[j] = list[j], list[i]
	}
	return list
}

// IDEA:
// when player sends message, check if it is his turn
// if it is not his turn, ignore his message
// only proceed the game when the right player sends a message
