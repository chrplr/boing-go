package main

import (
	"flag"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

// Screen and gameplay constants (untyped so they adapt to int or float64 use).
const (
	Width       = 800
	Height      = 480
	HalfWidth   = Width / 2
	HalfHeight  = Height / 2
	PlayerSpeed = 6
	MaxAISpeed  = 6
)

// normalised returns the unit vector pointing in the same direction as (x, y).
func normalised(x, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	return x / length, y / length
}

func abs(x float64) float64 { return math.Abs(x) }

type State int

const (
	StateMenu State = iota
	StatePlay
	StateGameOver
)

// Global game state, mirroring the module-level globals in the original.
var (
	state      State
	game       *Game
	numPlayers = 1
	spaceDown  bool

	assets *Assets
	audio  *Audio
)

func update() {
	// Detect a fresh press of a start key — Space or Enter — (down this frame, up the last).
	space := keyDown(sdl.SCANCODE_SPACE) || keyDown(sdl.SCANCODE_RETURN) || keyDown(sdl.SCANCODE_KP_ENTER)
	spacePressed := space && !spaceDown
	spaceDown = space

	switch state {
	case StateMenu:
		if spacePressed {
			// Start a game: player 1 is always human; player 2 is human in
			// two-player mode, otherwise AI (nil).
			controls := [2]func() float64{p1Controls, nil}
			if numPlayers == 2 {
				controls[1] = p2Controls
			}
			state = StatePlay
			game = NewGame(controls, assets, audio)
		} else {
			if numPlayers == 2 && keyDown(sdl.SCANCODE_UP) {
				audio.PlaySound("up", 1)
				numPlayers = 1
			} else if numPlayers == 1 && keyDown(sdl.SCANCODE_DOWN) {
				audio.PlaySound("down", 1)
				numPlayers = 2
			}
			// Run the AI-vs-AI attract-mode demo behind the menu.
			game.Update()
		}

	case StatePlay:
		if max(game.bats[0].score, game.bats[1].score) > 9 {
			state = StateGameOver
		} else {
			game.Update()
		}

	case StateGameOver:
		if spacePressed {
			state = StateMenu
			numPlayers = 1
			game = NewGame([2]func() float64{nil, nil}, assets, audio)
		}
	}
}

func draw() {
	game.Draw()
	switch state {
	case StateMenu:
		assets.Blit("menu"+itoa(numPlayers-1), 0, 0)
	case StateGameOver:
		assets.Blit("over", 0, 0)
	}
}

func main() {
	flag.Parse()

	a, err := pgzgo.New(pgzgo.Config{
		Title:  "Boing!",
		Width:  Width,
		Height: Height,
		Images: imagesFS,
		Audio:  audioFS,
	})
	if err != nil {
		panic(err)
	}
	defer a.Close()

	app = a
	assets = a.Screen
	audio = a.Audio

	state = StateMenu
	// The initial game has no players, so both bats are AI-controlled: this is the
	// attract-mode demo shown on the menu.
	game = NewGame([2]func() float64{nil, nil}, assets, audio)

	a.Loop(
		func(*pgzgo.App) { update() },
		func(*pgzgo.App) { draw() },
	)
}
