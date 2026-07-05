package main

import "github.com/Zyko0/go-sdl3/sdl"

// keyDown reports whether a key is held, via the harness keyboard snapshot.
func keyDown(sc sdl.Scancode) bool { return app.Keyboard.Held(sc) }

// p1Controls: Z or Down moves down, A or Up moves up.
func p1Controls() float64 {
	switch {
	case keyDown(sdl.SCANCODE_Z) || keyDown(sdl.SCANCODE_DOWN):
		return PlayerSpeed
	case keyDown(sdl.SCANCODE_A) || keyDown(sdl.SCANCODE_UP):
		return -PlayerSpeed
	}
	return 0
}

// p2Controls: M moves down, K moves up.
func p2Controls() float64 {
	switch {
	case keyDown(sdl.SCANCODE_M):
		return PlayerSpeed
	case keyDown(sdl.SCANCODE_K):
		return -PlayerSpeed
	}
	return 0
}
