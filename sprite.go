package main

// Sprite is a positioned image. Position (X, Y) is the sprite's centre, matching
// the behaviour of Pygame Zero's Actor, which anchors sprites from their centre.
type Sprite struct {
	X, Y  float64
	Image string
}

// Draw blits the sprite's current image centred on (X, Y).
func (s *Sprite) Draw(a *Assets) {
	a.BlitCentred(s.Image, s.X, s.Y)
}
