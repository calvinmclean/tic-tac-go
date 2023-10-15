package tictactoe

const (
	P1Piece = 'X'
	P2Piece = 'O'
)

type Player struct {
	GamePiece rune
	ID        string

	turn      chan bool
	gameOver  chan *bool
	update    chan Play
	errorChan chan string
}

func newPlayer(gamePiece rune, id string) *Player {
	return &Player{
		gamePiece, id,
		make(chan bool),
		make(chan *bool),
		make(chan Play),
		make(chan string),
	}
}

func (p *Player) NotifyGameOver() {
	go func() { p.gameOver <- nil }()
}

func (p *Player) NotifyWin() {
	t := true
	go func() { p.gameOver <- &t }()
}

func (p *Player) NotifyLoss() {
	f := false
	go func() { p.gameOver <- &f }()
}
