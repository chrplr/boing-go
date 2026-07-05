# Boing! — Python vs. Go implementation comparison

This document compares the Go port in this folder with the original `boing.py`.
Boing! is a two-player *Pong* clone with an AI opponent, an attract-mode demo on
the menu, and sub-stepped ball physics. It is the smallest of the ports, so it
is a good place to see the core translation patterns without much surrounding
machinery. The Go version is a faithful translation; the differences are
mechanical consequences of Go's static typing, the lack of class inheritance,
and swapping Pygame Zero for [go-sdl3](https://github.com/Zyko0/go-sdl3).

---

## 1. File organisation

Python is a single 475-line module; the Go port splits it into small files:

| Python section | Go file |
|---|---|
| `update()`, `draw()`, state machine, `p1/p2_controls`, mixer setup | `main.go` |
| `Game` class | `game.go` |
| `Ball` | `ball.go` |
| `Bat` (+ AI) | `bat.go` |
| `Impact` | `impact.go` |
| `Actor` behaviour | `sprite.go` |
| image blitting | `assets.go` |
| `play_sound` / music | `audio.go` |
| keyboard reading | `input.go` |

---

## 2. Inheritance → struct embedding

Python subclasses Pygame Zero's `Actor` three times:

```python
class Impact(Actor): ...
class Ball(Actor):   ...
class Bat(Actor):    ...
```

Go has no inheritance, so a small `Sprite` struct is embedded into each entity.
Because Pygame Zero anchors actors from their **centre**, `Sprite.Draw` blits
centred:

```go
type Sprite struct { X, Y float64; Image string }
func (s *Sprite) Draw(a *Assets) { a.BlitCentred(s.Image, s.X, s.Y) }

type Ball struct { Sprite; dx, dy float64; speed int }   // "Ball(Actor)"
type Bat  struct { Sprite; player, score, timer int; moveFunc func() float64 }
type Impact struct { Sprite; time int }
```

Each `update(self)` method becomes a method on the embedding struct. Where the
Python method needed the global `game`, the Go method takes `g *Game` as a
parameter (`func (b *Ball) Update(g *Game)`), avoiding a global.

---

## 3. First-class functions for controls and AI

Boing already uses functions as values in Python, and Go carries this over
directly.

- **Player controls** are passed into a bat as a callable:

  ```python
  def p1_controls(): ...          # returns -PLAYER_SPEED / 0 / PLAYER_SPEED
  game = Game([p1_controls, p2_controls if num_players == 2 else None])
  ```
  ```go
  func p1Controls() float64 { ... }
  controls := [2]func() float64{p1Controls, nil}
  if numPlayers == 2 { controls[1] = p2Controls }
  ```

- **`None` move function → AI.** In Python a bat with `move_func == None`
  substitutes `self.ai`; the "are we the menu demo?" test then compares
  `self.bats[0].move_func != self.bats[0].ai`. Go represents "no player function"
  as a `nil` func value and expresses the same logic with a helper:

  ```go
  func (b *Bat) isAI() bool { return b.moveFunc == nil }
  ...
  func (g *Game) PlaySound(name string, count int) {
      if g.bats[0].isAI() { return }   // muted during attract-mode demo
      g.audio.PlaySound(name, count)
  }
  ```

This is the neatest example in the whole project of `None` → `nil`: the callable
is optional, and its absence selects a different behaviour.

---

## 4. Vectors as scalar pairs

Boing's ball direction is a unit vector, but the Python code already stores it as
two scalars `self.dx, self.dy` and normalises with a helper. The Go port keeps
exactly this shape:

```python
def normalised(x, y):
    length = math.hypot(x, y)
    return (x / length, y / length)
...
self.dx, self.dy = normalised(self.dx, self.dy)
```
```go
func normalised(x, y float64) (float64, float64) {
    length := math.Hypot(x, y)
    return x / length, y / length
}
...
b.dx, b.dy = normalised(b.dx, b.dy)
```

Go's multiple return values map cleanly onto Python's tuple unpacking. The
sub-stepped physics loop (`for i in range(self.speed)`) is identical.

---

## 5. State enum → `const … iota`

```python
class State(Enum): MENU = 1; PLAY = 2; GAME_OVER = 3
```
```go
type State int
const ( StateMenu State = iota; StatePlay; StateGameOver )
```

The module-level globals (`state`, `game`, `num_players`, `space_down`) become
package-level `var`s, and the space-key edge detection is preserved verbatim:

```go
space := keyDown(sdl.SCANCODE_SPACE)
spacePressed := space && !spaceDown
spaceDown = space
```

---

## 6. List management idioms

Python removes finished impacts by iterating backwards and deleting in place:

```python
for i in range(len(self.impacts) - 1, -1, -1):
    if self.impacts[i].time >= 10:
        del self.impacts[i]
```

Go uses the idiomatic in-place filter that reuses the backing array:

```go
kept := g.impacts[:0]
for _, im := range g.impacts {
    if im.time < 10 { kept = append(kept, im) }
}
g.impacts = kept
```

Same result, but expressed with Go's slice-filtering idiom instead of reverse
deletion.

---

## 7. Text / sprite-name building

Both compose sprite names from numbers. Python uses f-strings and `getattr`; Go
uses `strconv`/`fmt` and a texture-cache lookup:

```python
score = f"{self.bats[p].score:02d}"      # e.g. "05"
image = "digit" + colour + str(score[i])
```
```go
score := fmt.Sprintf("%02d", g.bats[p].score)
image := "digit" + colour + string(score[i])   // score[i] is a byte, e.g. '0'
```

This port defines its own one-digit `itoa` (`string(rune('0'+i))`) for the
menu/effect indices — a different small helper than the other ports use, but the
same idea.

---

## 8. Framework specifics

| Concern | Python (Pygame Zero) | Go (go-sdl3) |
|---|---|---|
| Anchor | `Actor` centre anchor | `Sprite.Draw` → `Assets.BlitCentred` |
| Background/UI blit | `screen.blit("table", (0,0))` | `assets.Blit("table", 0, 0)` (top-left) |
| Input | `keyboard.z`, `keyboard.space`, … | per-frame `sdl.GetKeyboardState()` snapshot |
| Music | `music.play("theme")`, `set_volume(0.3)` | looping track in `Audio` |
| Loop | `pgzrun.go()` | explicit `sdl.RunLoop` with ~60 FPS cap |

Neither version uses a fixed-timestep accumulator: Pygame Zero runs the logic at
its frame rate, and the Go loop simply caps to ~60 FPS, matching it.

---

## 9. Audio and the attract-mode mute

`play_sound` behaviour is preserved, including the subtlety that **in-game
effects are muted while the menu's AI-vs-AI demo is running** (detected via the
player-0 bat being AI). The menu navigation blips (`up`/`down`) are played
unconditionally — in Python via the `menu_sound=True` flag; in Go by calling
`audio.PlaySound` directly rather than the game's muting wrapper:

```go
audio.PlaySound("up", 1)     // menu blip, always audible
...
g.PlaySound("hit", 5)        // gameplay, muted during the demo
```

---

## 10. What is intentionally identical

- Ball physics: sub-stepping by `speed`, the 344px bat-collision threshold with
  the "first step to cross" guard, the `difference_y / 128` deflection, unit-
  vector renormalisation, speed increase per hit, and wall bounces at 220px.
- Bat AI: the distance-weighted blend of "aim centre" vs. "aim at ball + random
  offset", clamped to `MAX_AI_SPEED`.
- Scoring: the loser's `timer` gating point award and the 20-frame delay before a
  new ball is served toward the player who missed.
- Menu/attract mode, 1P/2P selection, game-over at score > 9, and all sprite
  choices (bats, digits, effects, impacts).

---

## 11. Summary of differences

| Category | Difference | Reason |
|---|---|---|
| Inheritance | `Actor` subclasses → embedded `Sprite` | no classes |
| Optional callable | `move_func = None` → `moveFunc == nil` + `isAI()` | no `None` |
| Enums | `Enum` → `const … iota` | static typing |
| Vectors | already scalar `dx/dy` + `normalised` helper | (unchanged) |
| List removal | reverse `del` → in-place slice filter | Go idiom |
| Globals → params | global `game` → `g *Game` argument | avoids global coupling |
| Framework | Pygame Zero → go-sdl3 (blit, input, mixer, loop) | library swap |

Boing! is close to a one-to-one translation: the physics, AI, scoring, and state
machine are line-by-line equivalent, and the only genuinely interesting
translation is `None`-callable → `nil` func plus `isAI()`.
