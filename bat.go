package main

import "strconv"

// Bat is a paddle. If moveFunc is nil the bat is controlled by the AI; otherwise
// moveFunc returns the desired Y movement for this frame based on player input.
type Bat struct {
	Sprite
	player   int
	score    int
	timer    int
	moveFunc func() float64
}

func NewBat(player int, moveFunc func() float64) *Bat {
	x := 40.0
	if player != 0 {
		x = 760
	}
	return &Bat{
		Sprite:   Sprite{X: x, Y: HalfHeight, Image: "blank"},
		player:   player,
		moveFunc: moveFunc,
	}
}

// isAI reports whether this bat is computer-controlled.
func (b *Bat) isAI() bool { return b.moveFunc == nil }

func (b *Bat) Update(g *Game) {
	b.timer--

	var yMovement float64
	if b.isAI() {
		yMovement = b.ai(g)
	} else {
		yMovement = b.moveFunc()
	}

	// Apply the movement, keeping the bat within the walls.
	b.Y = min(400, max(80, b.Y+yMovement))

	// Pick the sprite: frame 1 when the bat has just hit the ball, frame 2 when
	// it has just missed and the ball is out of bounds, otherwise frame 0.
	frame := 0
	if b.timer > 0 {
		if g.ball.Out() {
			frame = 2
		} else {
			frame = 1
		}
	}
	b.Image = "bat" + strconv.Itoa(b.player) + strconv.Itoa(frame)
}

// ai returns how far the computer player should move this frame.
func (b *Bat) ai(g *Game) float64 {
	xDistance := abs(g.ball.X - b.X)

	// When the ball is far away, aim for the centre; when it is close, aim for the
	// ball's Y position plus a small random offset. Blend the two by distance so
	// the target sharpens as the ball approaches.
	targetY1 := float64(HalfHeight)
	targetY2 := g.ball.Y + float64(g.aiOffset)

	weight1 := min(1, xDistance/HalfWidth)
	weight2 := 1 - weight1
	targetY := weight1*targetY1 + weight2*targetY2

	return min(MaxAISpeed, max(-MaxAISpeed, targetY-b.Y))
}
