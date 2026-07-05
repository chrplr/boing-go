package main

import "strconv"

// Impact is a short animation played whenever the ball bounces. There are five
// impact sprites (impact0..impact4); we advance one every two frames, and the
// Game removes the impact once its timer reaches 10.
type Impact struct {
	Sprite
	time int
}

func NewImpact(x, y float64) *Impact {
	return &Impact{Sprite: Sprite{X: x, Y: y, Image: "blank"}}
}

func (im *Impact) Update() {
	im.Image = "impact" + strconv.Itoa(im.time/2)
	im.time++
}
