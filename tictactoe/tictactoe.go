package tictactoe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
)

const (
	emptySpace rune = '-'
	p1Key           = true
)

var (
	ErrNotYourTurn = errors.New("not your turn")
	ErrGameOver    = errors.New("game over")
)

type Game struct {
	ID      string
	players map[bool]*Player
	upNext  bool

	started  chan struct{}
	gameOver bool

	board [3][3]rune
}

type Handlers struct {
	OnPlay     func(play Play)
	OnTurn     func(turn bool)
	OnGameOver func(result *bool)
	OnErr      func(errMsg string)
}

type Play struct {
	Piece rune
	X, Y  int
}

func NewGame() *Game {
	g := &Game{
		ID:      random(8),
		players: map[bool]*Player{},
		started: make(chan struct{}),
	}
	g.upNext = p1Key

	for x := range g.board {
		for y := range g.board[x] {
			g.board[x][y] = emptySpace
		}
	}

	// Write to the channel so first player to join will know the game started
	go func() {
		g.started <- struct{}{}
	}()

	return g
}

func (g *Game) P1() *Player {
	return g.players[p1Key]
}

func (g *Game) P2() *Player {
	return g.players[!p1Key]
}

func (g *Game) NextPlayer() *Player {
	return g.players[g.upNext]
}

func (g *Game) Get(x, y int) rune {
	return g.board[x][y]
}

func (g *Game) Play(p *Player, x, y int) error {
	if g.gameOver {
		p.errorChan <- ErrGameOver.Error()
		return ErrGameOver
	}

	upNext := g.players[g.upNext]
	if p != upNext {
		p.errorChan <- ErrNotYourTurn.Error()
		return ErrNotYourTurn
	}

	if g.board[x][y] != emptySpace {
		err := fmt.Errorf("position (%d, %d) is taken", x, y)
		p.errorChan <- err.Error()
		return err
	}

	// All error checks are done, so we can notify that the turn is over
	p.turn <- false

	g.upNext = !g.upNext
	g.board[x][y] = p.GamePiece

	g.P1().update <- Play{p.GamePiece, x, y}
	g.P2().update <- Play{p.GamePiece, x, y}

	if g.BoardFull() {
		g.gameOver = true
		g.P1().NotifyGameOver()
		g.P2().NotifyGameOver()
		return nil
	}

	// Check if game over by win and notify
	switch g.WinFromPosition(x, y) {
	case p:
		g.gameOver = true
		p.NotifyWin()
		g.NextPlayer().NotifyLoss()
	case g.NextPlayer():
		g.gameOver = true
		g.NextPlayer().NotifyWin()
		p.NotifyLoss()
	case nil:
		if upNext := g.NextPlayer(); upNext != nil {
			upNext.turn <- true
		}
		return nil
	}

	return nil
}

func (g *Game) AddNewPlayer() *Player {
	return g.AddExistingPlayer(random(8))
}

func (g *Game) AddExistingPlayer(id string) *Player {
	p1 := g.P1() == nil
	var p *Player
	if p1 {
		p = newPlayer(P1Piece, id)
	} else {
		p = newPlayer(P2Piece, id)
	}

	log.Printf("Player %s added to the game\n", id)
	g.players[p1] = p

	return p
}

// Join will join the game and block
func (g *Game) Join(ctx context.Context, p *Player, funcs Handlers) {
	log.Printf("Player %s joined the game\n", p.ID)

	// Start the game if not started, otherwise continue to playing
	select {
	case <-g.started:
		log.Printf("Starting new game with %s as P1!", p.ID)
		go func() {
			p.turn <- true
		}()
	default:
	}

	for {
		select {
		case play := <-p.update:
			if funcs.OnPlay != nil {
				funcs.OnPlay(play)
			}

		case turn := <-p.turn:
			if funcs.OnTurn != nil {
				funcs.OnTurn(turn)
			}

		case result := <-p.gameOver:
			if funcs.OnGameOver != nil {
				funcs.OnGameOver(result)
			}

		case errMsg := <-p.errorChan:
			if funcs.OnErr != nil {
				funcs.OnErr(errMsg)
			}

		case <-ctx.Done():
			log.Printf("player %s disconnected", p.ID)
			return
		}
	}
}

func (g *Game) WinFromPosition(x, y int) *Player {
	var currentPlayer *Player
	switch g.Get(x, y) {
	case g.P1().GamePiece:
		currentPlayer = g.P1()
	case g.P2().GamePiece:
		currentPlayer = g.P2()
	}

	// Check horizontal
	if g.board[x][0]+g.board[x][1]+g.board[x][2] == currentPlayer.GamePiece*3 {
		return currentPlayer
	}

	// Check vertical
	if g.board[0][y]+g.board[1][y]+g.board[2][y] == currentPlayer.GamePiece*3 {
		return currentPlayer
	}

	// left -> right diagonal
	if (x == 1 && y == 1) || (x == 0 && y == 0) || (x == 2 && y == 2) {
		if g.board[0][0]+g.board[1][1]+g.board[2][2] == currentPlayer.GamePiece*3 {
			return currentPlayer
		}
	}

	// right -> left diagonal
	if (x == 1 && y == 1) || (x == 0 && y == 2) || (x == 2 && y == 0) {
		if g.board[0][2]+g.board[1][1]+g.board[2][0] == currentPlayer.GamePiece*3 {
			return currentPlayer
		}
	}

	return nil
}

func (g *Game) BoardFull() bool {
	for x := range g.board {
		for y := range g.board[x] {
			if g.board[x][y] == emptySpace {
				return false
			}
		}
	}

	return true
}

func (g *Game) String() string {
	return fmt.Sprintf(`   |   |
 %c | %c | %c
---|---|---
 %c | %c | %c
---|---|---
 %c | %c | %c
   |   |`,
		g.board[0][0], g.board[1][0], g.board[2][0],
		g.board[0][1], g.board[1][1], g.board[2][1],
		g.board[0][2], g.board[1][2], g.board[2][2],
	)
}

func (g *Game) GetPlayer(id string) *Player {
	if g.P1() != nil && g.P1().ID == id {
		return g.P1()
	}
	if g.P2() != nil && g.P2().ID == id {
		return g.P2()
	}

	return nil
}

func random(length int) string {
	characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = characters[rand.Intn(len(characters))]
	}

	return string(randomString)
}
