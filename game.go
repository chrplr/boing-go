package main

import "fmt"

// Game holds the two bats, the ball and any active impact animations. A single
// Game instance also drives the "attract mode" demo shown behind the menu.
type Game struct {
	bats     [2]*Bat
	ball     *Ball
	impacts  []*Impact
	aiOffset int

	assets *Assets
	audio  *Audio
}

// NewGame creates a game. Each entry of controls is either a player input
// function or nil for an AI-controlled bat.
func NewGame(controls [2]func() float64, assets *Assets, audio *Audio) *Game {
	g := &Game{assets: assets, audio: audio}
	g.bats[0] = NewBat(0, controls[0])
	g.bats[1] = NewBat(1, controls[1])
	g.ball = NewBall(-1)
	return g
}

func (g *Game) Update() {
	for _, bat := range g.bats {
		bat.Update(g)
	}
	g.ball.Update(g)
	for _, im := range g.impacts {
		im.Update()
	}

	// Drop any impact animations that have finished.
	kept := g.impacts[:0]
	for _, im := range g.impacts {
		if im.time < 10 {
			kept = append(kept, im)
		}
	}
	g.impacts = kept

	// Has the ball gone off the left or right edge?
	if g.ball.Out() {
		scoringPlayer := 0
		if g.ball.X < HalfWidth {
			scoringPlayer = 1
		}
		losingPlayer := 1 - scoringPlayer

		// The losing player's timer decides when a new ball appears. It counts
		// down each frame; the frame the ball goes out it is below zero, so we
		// award the point and set it to 20. After 20 frames we serve a new ball.
		if g.bats[losingPlayer].timer < 0 {
			g.bats[scoringPlayer].score++
			g.PlaySound("score_goal", 1)
			g.bats[losingPlayer].timer = 20
		} else if g.bats[losingPlayer].timer == 0 {
			direction := 1.0
			if losingPlayer == 0 {
				direction = -1
			}
			g.ball = NewBall(direction)
		}
	}
}

func (g *Game) Draw() {
	g.assets.Blit("table", 0, 0)

	// 'Just scored' effects.
	for p := 0; p < 2; p++ {
		if g.bats[p].timer > 0 && g.ball.Out() {
			g.assets.Blit("effect"+itoa(p), 0, 0)
		}
	}

	for _, bat := range g.bats {
		bat.Draw(g.assets)
	}
	g.ball.Draw(g.assets)
	for _, im := range g.impacts {
		im.Draw(g.assets)
	}

	// Scores: two digit sprites per player.
	for p := 0; p < 2; p++ {
		score := fmt.Sprintf("%02d", g.bats[p].score)
		for i := 0; i < 2; i++ {
			// Digit sprites are digit<colour><n>: colour 0 grey, 1 blue, 2 green.
			// The scorer's digits flash their colour while the ball is out.
			colour := "0"
			otherP := 1 - p
			if g.bats[otherP].timer > 0 && g.ball.Out() {
				if p == 0 {
					colour = "2"
				} else {
					colour = "1"
				}
			}
			image := "digit" + colour + string(score[i])
			g.assets.Blit(image, float64(255+160*p+i*55), 46)
		}
	}
}

// PlaySound plays an in-game sound effect. As in the original, effects are muted
// when player 0 is an AI (i.e. we are on the menu's attract-mode demo).
func (g *Game) PlaySound(name string, count int) {
	if g.bats[0].isAI() {
		return
	}
	g.audio.PlaySound(name, count)
}

func itoa(i int) string {
	return string(rune('0' + i))
}
