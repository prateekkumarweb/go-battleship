package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Settings
var gridRows int = 10
var gridCols int = 10
var ships []int = []int{5, 4, 3, 3, 2}

// Game Status
var inProgress int = 1
var gameOver int = 2

// Ship
type Ship struct {
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Size       int  `json:"size"`
	Hits       int  `json:"hits"`
	Horizontal bool `json:"horizontal"`
}

func (s *Ship) init(sz int) {
	s.X = 0
	s.Y = 0
	s.Size = sz
	s.Hits = 0
	s.Horizontal = false
}

func (s *Ship) isSunk() bool {
	return s.Hits >= s.Size
}

// Player
type Player struct {
	Id       string
	Shots    []int
	ShipGrid []int
	Ships    []Ship
}

func (p *Player) init(id string) {
	p.Id = id
	p.Shots = make([]int, gridRows*gridCols)
	p.ShipGrid = make([]int, gridRows*gridCols)
	p.Ships = make([]Ship, 0)

	for i := 0; i < gridRows*gridCols; i++ {
		p.Shots[i] = 0
		p.ShipGrid[i] = -1
	}

	if !p.createRandomShips() {
		p.Ships = make([]Ship, 0)
		p.createShips()
	}
}

func (p *Player) shoot(gridIndex int) bool {
	if p.ShipGrid[gridIndex] >= 0 {
		p.Ships[p.ShipGrid[gridIndex]].Hits++
		p.Shots[gridIndex] = 2
		return true
	} else {
		p.Shots[gridIndex] = 1
		return false
	}
}

func (p *Player) getSunkShips() []Ship {
	sunkShips := make([]Ship, 0)
	for i := 0; i < len(p.Ships); i++ {
		if p.Ships[i].isSunk() {
			sunkShips = append(sunkShips, p.Ships[i])
		}
	}
	return sunkShips
}

func (p *Player) getShipsLeft() int {
	shipCount := 0
	for i := 0; i < len(p.Ships); i++ {
		if !p.Ships[i].isSunk() {
			shipCount++
		}
	}
	return shipCount
}

func (p *Player) createRandomShips() bool {
	for shipIndex := 0; shipIndex < len(ships); shipIndex++ {
		var ship Ship
		ship.init(ships[shipIndex])
		if !p.placeShipRandom(&ship, shipIndex) {
			return false
		}
		p.Ships = append(p.Ships, ship)
	}
	return true
}

func (p *Player) placeShipRandom(ship *Ship, shipIndex int) bool {
	tryMax := 25
	var xMax int
	var yMax int
	for i := 0; i < tryMax; i++ {
		ship.Horizontal = rand.Float64() < 0.5
		if ship.Horizontal {
			xMax = gridCols - ship.Size + 1
			yMax = gridRows
		} else {
			xMax = gridCols
			yMax = gridRows - ship.Size + 1
		}

		ship.X = int(math.Floor(rand.Float64() * float64(xMax)))
		ship.Y = int(math.Floor(rand.Float64() * float64(yMax)))
		if !p.checkShipOverlap(*ship) && !p.checkShipAdjacent(*ship) {
			gridIndex := ship.Y*gridCols + ship.X
			for j := 0; j < ship.Size; j++ {
				p.ShipGrid[gridIndex] = shipIndex
				if ship.Horizontal {
					gridIndex += 1
				} else {
					gridIndex += gridCols
				}
			}
			return true
		}
	}
	return false
}

func (p *Player) checkShipOverlap(ship Ship) bool {
	gridIndex := ship.Y*gridCols + ship.X
	for i := 0; i < ship.Size; i++ {
		if p.ShipGrid[gridIndex] >= 0 {
			return true
		}
		if ship.Horizontal {
			gridIndex += 1
		} else {
			gridIndex += gridCols
		}
	}
	return false
}

func (p *Player) checkShipAdjacent(ship Ship) bool {
	x1 := ship.X - 1
	y1 := ship.Y - 1
	var x2, y2 int
	if ship.Horizontal {
		x2 = ship.X + ship.Size
		y2 = ship.Y + 1
	} else {
		x2 = ship.X + 1
		y2 = ship.Y + ship.Size
	}

	for i := x1; i <= x2; i++ {
		if i < 0 || i > gridCols-1 {
			continue
		}
		for j := y1; j <= y2; j++ {
			if j < 0 || j > gridRows-1 {
				continue
			}
			if p.ShipGrid[j*gridCols+i] >= 0 {
				return true
			}
		}
	}
	return false
}

func (p *Player) createShips() {
	x := []int{1, 3, 5, 8, 8}
	y := []int{1, 2, 5, 2, 8}
	horizontal := []bool{false, true, false, false, true}

	for shipIndex := 0; shipIndex < len(ships); shipIndex++ {
		var ship Ship
		ship.init(ships[shipIndex])
		ship.Horizontal = horizontal[shipIndex]
		ship.X = x[shipIndex]
		ship.Y = y[shipIndex]

		gridIndex := ship.Y*gridCols + ship.X
		for i := 0; i < ship.Size; i++ {
			p.ShipGrid[gridIndex] = shipIndex
			if ship.Horizontal {
				gridIndex += 1
			} else {
				gridIndex += gridCols
			}
		}
		p.Ships = append(p.Ships, ship)
	}
}

// Game
type Game struct {
	Id            int
	CurrentPlayer int
	WinningPlayer int
	GameStatus    int
	Players       []Player
}

func (g *Game) init(id int, idPlayer1 string, idPlayer2 string) {
	g.Id = id
	g.CurrentPlayer = int(math.Floor(rand.Float64() * 2))
	g.WinningPlayer = -1 // null
	g.GameStatus = inProgress
	var p1, p2 Player
	p1.init(idPlayer1)
	p2.init(idPlayer2)
	g.Players = []Player{p1, p2}
}

func (g *Game) getPlayerId(player int) string {
	return g.Players[player].Id
}

func (g *Game) getWinnerId() string {
	if g.WinningPlayer == -1 {
		return ""
	}
	return g.Players[g.WinningPlayer].Id
}

func (g *Game) getLoserId() string {
	if g.WinningPlayer == -1 {
		return ""
	}
	var loser int
	if g.WinningPlayer == 0 {
		loser = 1
	} else {
		loser = 0
	}
	return g.Players[loser].Id
}

func (g *Game) switchPlayer() {
	if g.CurrentPlayer == 0 {
		g.CurrentPlayer = 1
	} else {
		g.CurrentPlayer = 0
	}
}

func (g *Game) abortGame(player int) {
	g.GameStatus = gameOver
	if player == 0 {
		g.WinningPlayer = 1
	} else {
		g.WinningPlayer = 0
	}
}

func (g *Game) shoot(x int, y int) bool {
	var opponent int
	if g.CurrentPlayer == 0 {
		opponent = 1
	} else {
		opponent = 0
	}
	gridIndex := y*gridCols + x

	if g.Players[opponent].Shots[gridIndex] == 0 && g.GameStatus == inProgress {
		if !g.Players[opponent].shoot(gridIndex) {
			g.switchPlayer()
		}
		if g.Players[opponent].getShipsLeft() <= 0 {
			g.GameStatus = gameOver
			if opponent == 0 {
				g.WinningPlayer = 1
			} else {
				g.WinningPlayer = 0
			}
		}
		return true
	}
	return false
}

type GameState struct {
	Turn      bool `json:"turn"`
	GridIndex int  `json:"gridIndex"`
	Grid      Grid `json:"grid"`
}

func (g *Game) getGameState(player int, gridOwner int) GameState {
	turn := g.CurrentPlayer == player
	var gridIndex int
	if player == gridOwner {
		gridIndex = 0
	} else {
		gridIndex = 1
	}
	grid := g.getGrid(gridOwner, player != gridOwner)
	return GameState{turn, gridIndex, grid}
}

type Grid struct {
	Shots []int  `json:"shots"`
	Ships []Ship `json:"ships"`
}

func (g *Game) getGrid(player int, hideShips bool) Grid {
	shts := g.Players[player].Shots
	var shps []Ship
	if hideShips {
		shps = g.Players[player].getSunkShips()
	} else {
		shps = g.Players[player].Ships
	}
	return Grid{shts, shps}
}

var players map[string]*websocket.Conn
var users map[string]struct {
	inGame *Game
	player int
}
var waiting []string
var gameIdCounter int = 1

func main() {
	port := flag.String("http", ":8000", "Port on which server should run")
	flag.Parse()

	players = make(map[string]*websocket.Conn)
	users = make(map[string]struct {
		inGame *Game
		player int
	})
	waiting = make([]string, 0)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		go HandleClient(conn)
	})

	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.ListenAndServe(*port, nil)
}

type Message struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

func HandleClient(c *websocket.Conn) {
	id := uuid.NewV4().String()
	fmt.Printf("%s ID %s connected.\n", time.Now().String(), id)

	players[id] = c
	users[id] = struct {
		inGame *Game
		player int
	}{nil, -1}

	waiting = append(waiting, id)

	joinWaitingPlayers()

	defer c.Close()
	// msgs := make(chan string)

	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			handleMessage(id, Message{"disconnect", ""})
			return
		}
		handleMessage(id, msg)
	}
}

func handleMessage(id string, msg Message) {
	c := players[id]
	if msg.Event == "chat" {
		if users[id].inGame != nil && msg.Message != "" {
			fmt.Printf("%s Chat message from %s : "+msg.Message, time.Now().String(), id)
			var opp *websocket.Conn
			opp0 := users[id].inGame.Players[0].Id
			opp1 := users[id].inGame.Players[1].Id
			if players[opp0] == c {
				opp = players[opp1]
			} else {
				opp = players[opp0]
			}

			jsonb, _ := json.Marshal(struct {
				Name    string `json:"name"`
				Message string `json:"message"` // encode ?
			}{"Opponent", msg.Message})

			opp.WriteJSON(Message{"chat", string(jsonb)})

			jsonb, _ = json.Marshal(struct {
				Name    string `json:"name"`
				Message string `json:"message"`
			}{"Me", msg.Message})

			c.WriteJSON(Message{"chat", string(jsonb)})
		}
	} else if msg.Event == "shot" {
		game := users[id].inGame
		if game != nil {
			if game.CurrentPlayer == users[id].player {
				var opponent int
				if game.CurrentPlayer == 0 {
					opponent = 1
				} else {
					opponent = 0
				}
				var position struct {
					X int `json:"x"`
					Y int `json:"y"`
				}
				json.Unmarshal([]byte(msg.Message), &position)
				if game.shoot(position.X, position.Y) {
					checkGameOver(game)
					jsonb, _ := json.Marshal(game.getGameState(users[id].player, opponent))
					players[id].WriteJSON(Message{"update", string(jsonb)})
					jsonb, _ = json.Marshal(game.getGameState(opponent, opponent))
					players[game.getPlayerId(opponent)].WriteJSON(Message{"update", string(jsonb)})
				}
			}
		}
	} else if msg.Event == "leave" {
		if users[id].inGame != nil {
			leaveGame(id)
			waiting = append(waiting, id)
			joinWaitingPlayers()
		}
	} else if msg.Event == "disconnect" {
		fmt.Printf("%s ID %s disconnected.\n", time.Now().String(), id)
		leaveGame(id)
		delete(users, id)
		delete(players, id)
	}
}

func joinWaitingPlayers() {
	if len(waiting) >= 2 {
		p1 := waiting[0]
		p2 := waiting[1]
		waiting = waiting[2:]
		var game Game
		game.init(gameIdCounter, p1, p2)
		gameIdCounter++

		users[p1] = struct {
			inGame *Game
			player int
		}{&game, 0}
		users[p2] = struct {
			inGame *Game
			player int
		}{&game, 1}

		players[p1].WriteJSON(Message{"join", strconv.Itoa(game.Id)})
		players[p2].WriteJSON(Message{"join", strconv.Itoa(game.Id)})

		jsonb, _ := json.Marshal(game.getGameState(0, 0))
		players[p1].WriteJSON(Message{"update", string(jsonb)})
		jsonb, _ = json.Marshal(game.getGameState(1, 1))
		players[p2].WriteJSON(Message{"update", string(jsonb)})

		fmt.Printf("%s %s and %s have joined game ID %d\n", time.Now().String(), p1, p2, game.Id)
	}
}

func leaveGame(id string) {
	if users[id].inGame != nil {
		fmt.Printf("%s ID %s left game ID %s", time.Now().String(), id, users[id].inGame.Id)

		jsonb, _ := json.Marshal(struct{Message string `json:"message"`}{"Opponent has left the game"})
		if users[id].player == 0 {
			players[users[id].inGame.Players[1].Id].WriteJSON(Message{"notification", string(jsonb)})
		} else {
			players[users[id].inGame.Players[0].Id].WriteJSON(Message{"notification", string(jsonb)})
		}

		if users[id].inGame.GameStatus != gameOver {
			users[id].inGame.abortGame(users[id].player)
			checkGameOver(users[id].inGame)
		}
		users[id] = struct {
			inGame *Game
			player int
		}{nil, -1}
		players[id].WriteJSON(Message{"leave", ""})
	}
}

func checkGameOver(game *Game) {
	if game.GameStatus == gameOver {
		fmt.Printf("%s Game ID %d ended.\n", time.Now().String(), game.Id)
		p1 := game.getWinnerId()
		p2 := game.getLoserId()
		jsonb, _ := json.Marshal(struct {
			IsWinner bool `json:"isWinner"`
		}{true})
		players[p1].WriteJSON(Message{"gameover", string(jsonb)})
		jsonb, _ = json.Marshal(struct {
			IsWinner bool `json:"isWinner"`
		}{false})
		players[p2].WriteJSON(Message{"gameover", string(jsonb)})
	}
}
