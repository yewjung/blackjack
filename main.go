package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/thanhpk/randstr"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := []string{"http://127.0.0.1:5173"}
		for _, origin := range allowedOrigins {
			if r.Header.Get("Origin") == origin {
				return true
			}
		}
		return false
	},
}

type Player struct {
	ID       string
	Conn     *websocket.Conn
	RoomId   string
	GameData GameData
}

type Request struct {
	Action Action `json:"action"`
	RoomId string `json:"roomId"`
}

type Response struct {
	Event        Event  `json:"event"`
	RoomId       string `json:"roomId,omitempty"`
	Error        *Error `json:"error,omitempty"`
	PlayerId     string `json:"playerId,omitempty"`
	NextPlayer   string `json:"nextPlayer,omitempty"`
	AffectedUser string `json:"affectedUser,omitempty"`
	NewCard      *Card  `json:"newCard,omitempty"`
}

type Room struct {
	PlayerMap map[string]*Player
	Game      Game
}

var players = make(map[string]*Player)
var rooms = make(map[string]Room)

func main() {
	fmt.Println("server start")
	http.HandleFunc("/ws", handleWebsocket)

	// for debugging purposes
	// TODO: remove this after debugging
	http.HandleFunc("/players", handlePlayers)
	http.HandleFunc("/rooms", handleRooms)

	http.ListenAndServe(":8080", nil)
}

func handlePlayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}
func handleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	player := createPlayer(conn)
	defer func() {
		delete(players, player.ID)
		leaveRoom(player)
	}()
	// send appropriate response back to client
	playerCreated := sendPlayerAddingResponse(player, conn)
	if !playerCreated {
		return
	}

	for {
		// messageType is either BinaryMessage(2) or TextMessage(1)
		_, message, err := conn.ReadMessage()
		if err != nil {
			delete(players, player.ID)
			leaveRoom(player)
			return
		}

		var request Request
		json.Unmarshal(message, &request)

		switch request.Action {
		case JOIN_ROOM:
			joinRoom(player, request.RoomId)
		case CREATE_ROOM:
			createRoom(player)
		case LEAVE_ROOM:
			leaveRoom(player)
		case START_GAME:
			startBlackjackGame(player)
		case SEND_HIT:
			room := rooms[player.RoomId]
			room.Game.DrawCard(player)
		case SEND_STAND:
			room := rooms[player.RoomId]
			room.Game.Stand(player)
		}
	}
}

/*
Returns true if player was created
*/
func sendPlayerAddingResponse(player *Player, conn *websocket.Conn) bool {
	playerAddedEvent := Response{
		Event:    PLAYER_ADDED,
		PlayerId: player.ID,
	}
	conn.WriteJSON(playerAddedEvent)
	return true
}

/*
Creates a player and add him into memory
Returns nil if player wasn't created
*/
func createPlayer(conn *websocket.Conn) *Player {
	player := &Player{
		ID:   generatePlayerID(),
		Conn: conn,
	}
	players[player.ID] = player
	return player
}

func createRoom(player *Player) {
	// player cannot create room if he is already in another room
	if _, ok := rooms[player.RoomId]; ok {
		log.Printf("[%s]: Cannot create room while already in another room", player.ID)
		response := createErrorResponse(CANNOT_CREATE_ROOM)
		player.Conn.WriteJSON(response)
		return
	}
	// player is either in an invalid room or no room
	// either case, roomId should be empty
	player.RoomId = ""

	// create a new room
	roomId := randstr.String(6)
	_, exist := rooms[roomId]
	for exist {
		roomId = randstr.String(6)
		_, exist = rooms[roomId]
	}
	// insert new room into rooms
	playerMap := map[string]*Player{
		player.ID: player,
	}
	newRoom := Room{
		PlayerMap: playerMap,
	}
	rooms[roomId] = newRoom

	player.RoomId = roomId

	// send back room id to client
	response := Response{
		Event:  ROOM_CREATED,
		RoomId: roomId,
	}
	player.Conn.WriteJSON(response)

}

func createErrorResponse(err Error) Response {
	response := Response{
		Event: ERROR,
		Error: &err,
	}
	return response
}

func generatePlayerID() string {
	u := uuid.NewV4()
	return u.String()
}

func joinRoom(player *Player, roomIDUncast interface{}) {
	// player cannot joinRoom while in another room
	currentRoomId := player.RoomId
	if _, ok := rooms[currentRoomId]; ok {
		log.Printf("[%s]: Cannot join room while in another room", player.ID)
		response := createErrorResponse(CANNOT_JOIN_ROOM)
		player.Conn.WriteJSON(response)
		return
	}
	if roomIDUncast == nil {
		log.Printf("[%s]: roomId is nil. Cannot be casted", player.ID)
		response := createErrorResponse(ROOM_ID_NOT_PROVIDED)
		player.Conn.WriteJSON(response)
		return
	}
	roomId, ok := roomIDUncast.(string)
	if !ok {
		log.Printf("[%s]: roomId casting to string failed", player.ID)
		response := createErrorResponse(ROOM_ID_FORMAT_WRONG)
		player.Conn.WriteJSON(response)
		return
	}

	if _, ok := rooms[roomId]; !ok {
		log.Printf("[%s]: room with id=%s doesn't exist", player.ID, roomId)
		response := createErrorResponse(ROOM_NOT_EXIST)
		player.Conn.WriteJSON(response)
		return
	}

	// adding new player into the room
	player.RoomId = roomId
	room := rooms[roomId]
	room.PlayerMap[player.ID] = player
	rooms[roomId] = room

	if len(rooms[roomId].PlayerMap) >= 2 {
		startGame(roomId)
	}
}

func leaveRoom(player *Player) {
	roomId := player.RoomId
	room, ok := rooms[roomId]
	if ok {
		delete(room.PlayerMap, player.ID)
	}
	player.RoomId = ""

	// delete room if it is empty
	if len(room.PlayerMap) == 0 {
		delete(rooms, roomId)
	}
}

func startGame(roomId string) {
	players := rooms[roomId].PlayerMap
	game := Game{Deck: []Card{}, DealerHands: []Card{}, Players: []*Player{}}
	for _, player := range players {
		game.Players = append(game.Players, player)
	}
	game.StartGame()
}

func startBlackjackGame(player *Player) {
	fmt.Println("Starting Blackjack game...")

	player.Conn.WriteJSON("game in progress")
}
