package tictactoe

import (
	"context"
	"testing"
)

func TestTicTacToeWin(t *testing.T) {
	type move struct {
		x, y int
	}

	tests := []struct {
		name   string
		moves  []move
		winP1  bool
		nilWin bool
	}{
		{
			"NoWin",
			[]move{
				{0, 0},
			},
			false, true,
		},
		{
			"P1WinsHorizontalFirstRow",
			[]move{
				{0, 0},
				{1, 0},
				{0, 1},
				{1, 1},
				{0, 2},
			},
			true, false,
		},
		{
			"P1WinsHorizontalSecondRow",
			[]move{
				{1, 0},
				{0, 0},
				{1, 1},
				{0, 1},
				{1, 2},
			},
			true, false,
		},
		{
			"P1WinsHorizontalThirdRow",
			[]move{
				{2, 0},
				{0, 0},
				{2, 1},
				{0, 1},
				{2, 2},
			},
			true, false,
		},
		{
			"P1WinsVerticalFirstCol",
			[]move{
				{0, 0},
				{0, 1},
				{1, 0},
				{0, 2},
				{2, 0},
			},
			true, false,
		},
		{
			"P1WinsLeftRightDiagonal",
			[]move{
				{0, 0},
				{0, 1},
				{1, 1},
				{0, 2},
				{2, 2},
			},
			true, false,
		},
		{
			"P1WinsLeftRightDiagonalWithFinalPieceInMiddle",
			[]move{
				{0, 0},
				{0, 1},
				{2, 2},
				{0, 2},
				{1, 1},
			},
			true, false,
		},
		{
			"P1WinsRightLeftDiagonal",
			[]move{
				{0, 2},
				{0, 1},
				{1, 1},
				{1, 2},
				{2, 0},
			},
			true, false,
		},
		{
			"P1WinsRightLeftDiagonalWithFinalPieceInMiddle",
			[]move{
				{0, 2},
				{0, 1},
				{2, 0},
				{1, 2},
				{1, 1},
			},
			true, false,
		},
		{
			"P2WinsHorizontalFirstRow",
			[]move{
				{1, 0},
				{0, 0},
				{1, 1},
				{0, 1},
				{2, 2},
				{0, 2},
			},
			false, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			p1 := game.AddNewPlayer()
			p2 := game.AddNewPlayer()

			go game.Join(context.Background(), p1, Handlers{})
			go game.Join(context.Background(), p2, Handlers{})

			for i, move := range tt.moves {
				var p *Player
				switch i % 2 {
				case 0:
					p = p1
				case 1:
					p = p2
				}
				err := game.Play(p, move.x, move.y)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			winner := game.WinFromPosition(tt.moves[len(tt.moves)-1].x, tt.moves[len(tt.moves)-1].y)
			if tt.nilWin && winner != nil {
				t.Errorf("expected nil winner but got: %+v", winner)
			}
			if tt.winP1 && winner != p1 {
				t.Errorf("expected p1 to win, but got: %+v", winner)
			}
		})
	}
}
