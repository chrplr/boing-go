package main

import "math/rand"

// Ball moves each frame in a series of small unit steps. dx and dy form a unit
// vector describing its direction; speed is the number of steps taken per frame.
type Ball struct {
	Sprite
	dx, dy float64
	speed  int
}

// NewBall creates a ball at the centre of the screen heading horizontally in the
// direction given by dx (-1 = left, 1 = right).
func NewBall(dx float64) *Ball {
	return &Ball{
		Sprite: Sprite{X: HalfWidth, Y: HalfHeight, Image: "ball"},
		dx:     dx,
		dy:     0,
		speed:  5,
	}
}

// Out reports whether the ball has left the left or right edge of the screen.
func (b *Ball) Out() bool {
	return b.X < 0 || b.X > Width
}

func (b *Ball) Update(g *Game) {
	// Move in a series of small steps, the count based on the ball's speed.
	for i := 0; i < b.speed; i++ {
		originalX := b.X

		b.X += b.dx
		b.Y += b.dy

		// A bat's centre is 360px from the centre of the screen; accounting for
		// the half-widths of bat (9) and ball (7), the ball can strike a bat once
		// its centre is 344px from the middle. We also require that this is the
		// first step to cross the threshold.
		if abs(b.X-HalfWidth) >= 344 && abs(originalX-HalfWidth) < 344 {
			var newDirX float64
			var bat *Bat
			if b.X < HalfWidth {
				newDirX = 1
				bat = g.bats[0]
			} else {
				newDirX = -1
				bat = g.bats[1]
			}

			differenceY := b.Y - bat.Y
			if differenceY > -64 && differenceY < 64 {
				// Bounce back on the X axis, and deflect slightly up or down based
				// on where the ball struck the bat, giving the player some control.
				b.dx = -b.dx
				b.dy += differenceY / 128
				b.dy = min(max(b.dy, -1), 1)
				b.dx, b.dy = normalised(b.dx, b.dy)

				g.impacts = append(g.impacts, NewImpact(b.X-newDirX*10, b.Y))

				b.speed++
				g.aiOffset = rand.Intn(21) - 10
				bat.timer = 10

				// Play hit sounds, more intense as the ball speeds up.
				g.PlaySound("hit", 5)
				switch {
				case b.speed <= 10:
					g.PlaySound("hit_slow", 1)
				case b.speed <= 12:
					g.PlaySound("hit_medium", 1)
				case b.speed <= 16:
					g.PlaySound("hit_fast", 1)
				default:
					g.PlaySound("hit_veryfast", 1)
				}
			}
		}

		// Bounce off the top and bottom walls, 220px from the centre.
		if abs(b.Y-HalfHeight) > 220 {
			b.dy = -b.dy
			b.Y += b.dy

			g.impacts = append(g.impacts, NewImpact(b.X, b.Y))

			g.PlaySound("bounce", 5)
			g.PlaySound("bounce_synth", 1)
		}
	}
}
