package main

// EXTRA FEATURES
//   - "New Game" button will start a new game and P2 becomes P1

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"github.com/calvinmclean/tic-tac-go/tictactoe"
)

const (
	gameIDQueryParam = "gameID"
)

type GameServer struct {
	games map[string]*tictactoe.Game
}

func NewGameServer() *GameServer {
	return &GameServer{map[string]*tictactoe.Game{}}
}

func getPlayerIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("Player")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		fmt.Println("err get cookie:", err)
		return "", fmt.Errorf("error getting cookie: %w", err)
	}
	if cookie == nil {
		return "", nil
	}
	return cookie.Value, nil
}

func (g *GameServer) joinOrCreateGame(w http.ResponseWriter, r *http.Request) {
	// Create game gameID if it is not included
	gameID := r.URL.Query().Get(gameIDQueryParam)
	if gameID == "" {
		game := tictactoe.NewGame()
		gameID = game.ID
		g.games[gameID] = game

		http.Redirect(w, r, fmt.Sprintf("?%s=%s", gameIDQueryParam, gameID), http.StatusFound)
		return
	}

	game := g.games[gameID]

	// Create player if not created
	playerID, err := getPlayerIDFromCookie(r)
	if err != nil {
		fmt.Println("error getting player ID from cookie to join or create:", err)
		return
	}

	if playerID == "" {
		p := game.AddNewPlayer()

		cookie := http.Cookie{
			Name:     "Player",
			Value:    url.QueryEscape(p.ID),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(10 * time.Minute),
		}
		http.SetCookie(w, &cookie)
	} else if game.GetPlayer(playerID) == nil {
		game.AddExistingPlayer(playerID)
	}

	fmt.Println("joining game with ID", gameID)
	fmt.Print(game)

	g.renderHTMX(w, gameID)
}

func (g *GameServer) startServerSideEvents(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get handshake from client")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	gameID := r.URL.Query().Get(gameIDQueryParam)
	game := g.games[gameID]
	if game == nil {
		fmt.Println("no game with ID", gameID)
		return
	}

	playerID, err := getPlayerIDFromCookie(r)
	if err != nil {
		fmt.Println("error getting player ID from cookie to start SSE:", err)
		return
	}

	if playerID == "" {
		fmt.Println("missing player ID:", err)
		return
	}

	player := game.GetPlayer(playerID)
	if player == nil {
		player = game.AddNewPlayer()
	}
	game.Join(r.Context(), player, tictactoe.Handlers{
		OnPlay: func(play tictactoe.Play) {
			fmt.Fprintf(w, "event: event%d%d\n", play.X, play.Y)
			fmt.Fprintf(w, "data: %c\n\n", play.Piece)
			w.(http.Flusher).Flush()
		},
		OnTurn: func(turn bool) {
			fmt.Fprintf(w, "event: eventTurnNotifier\n")
			not := "not "
			if turn {
				not = ""
			}
			fmt.Fprintf(w, "data: %syour turn!\n\n", not)
			w.(http.Flusher).Flush()
		},
		OnGameOver: func(result *bool) {
			fmt.Fprintf(w, "event: eventGameOver\n")
			msg := "game over!"

			if result != nil {
				switch *result {
				case true:
					msg = "You Win!"
				case false:
					msg = "You Lose!"
				}
			}

			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		},
		OnErr: func(errMsg string) {
			fmt.Fprintf(w, "event: eventError\n")
			fmt.Fprintf(w, "data: %s\n\n", errMsg)

			// Reset error after 2 seconds
			time.AfterFunc(2*time.Second, func() {
				fmt.Fprintf(w, "event: eventError\n")
				fmt.Fprintf(w, "data: \n\n")
				w.(http.Flusher).Flush()
			})
			w.(http.Flusher).Flush()
		},
	})

	log.Printf("player disconnected")
}

func (g *GameServer) makeMove(w http.ResponseWriter, r *http.Request, game *tictactoe.Game) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	xStr := r.FormValue("x")
	yStr := r.FormValue("y")
	playerID, err := getPlayerIDFromCookie(r)
	if err != nil {
		fmt.Println("error getting player ID from cookie to make move:", err)
		return
	}

	x, err := strconv.Atoi(xStr)
	if err != nil {
		http.Error(w, "Invalid value for 'x'", http.StatusBadRequest)
		return
	}

	y, err := strconv.Atoi(yStr)
	if err != nil {
		http.Error(w, "Invalid value for 'y'", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received coordinates from Player %s: X=%d, Y=%d\n", playerID, x, y)

	p := game.GetPlayer(playerID)
	err = game.Play(p, x, y)
	if err != nil {
		fmt.Printf("error playing: %v", err)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{byte(game.Get(x, y))})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{byte(p.GamePiece)})
}

func (g *GameServer) play(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get(gameIDQueryParam)
	game := g.games[gameID]

	switch r.Method {
	case http.MethodGet:
		g.startServerSideEvents(w, r)
		return
	case http.MethodPost:
		g.makeMove(w, r, game)
	}
}

type Page struct {
	Grid   [][]int
	GameID string
}

func (g *GameServer) renderHTMX(w http.ResponseWriter, gameID string) {
	game := g.games[gameID]

	data := Page{
		Grid: [][]int{
			{0, 1, 2},
			{0, 1, 2},
			{0, 1, 2},
		},
		GameID: gameID,
	}

	tmpl, err := template.
		New("tictactoe.html").
		Funcs(template.FuncMap{
			"GetPiece": func(x, y int) string {
				return string(game.Get(x, y))
			},
		}).
		ParseFiles("tictactoe.html")
	if err != nil {
		fmt.Println("err parse:", err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("err execute:", err)
		return
	}
}

func main() {
	server := NewGameServer()

	http.HandleFunc("/", server.joinOrCreateGame)
	http.HandleFunc("/tictactoe", server.play)

	fmt.Println("running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
