package main

import (
	"fmt"

	"github.com/qeedquan/go-media/sdl"
)

type Fruit struct {
	pos       sdl.Point
	vel       sdl.Point
	nearest   sdl.Point
	active    bool
	images    []*Image
	typ       int
	speed     int
	bouncei   int
	bounceY   int
	slowTimer int
	path      string
	hasPath   bool
}

func newFruit() *Fruit {
	var images []*Image
	for i := 0; i < 5; i++ {
		images = append(images, loadImage(fmt.Sprintf("sprite/fruit %d.gif", i)))
	}

	return &Fruit{
		images:  images,
		pos:     sdl.Point{-Tile.X, -Tile.Y},
		nearest: sdl.Point{-1, -1},
		speed:   2,
		typ:     1,
	}
}

func (f *Fruit) Move() {
	if !f.active {
		return
	}

	switch f.bouncei++; f.bouncei {
	case 1:
		f.bounceY = 2
	case 2:
		f.bounceY = 4
	case 3:
		f.bounceY = 5
	case 4:
		f.bounceY = 5
	case 5:
		f.bounceY = 6
	case 6:
		f.bounceY = 6
	case 9:
		f.bounceY = 6
	case 10:
		f.bounceY = 5
	case 11:
		f.bounceY = 5
	case 12:
		f.bounceY = 4
	case 13:
		f.bounceY = 3
	case 14:
		f.bounceY = 2
	case 15:
		f.bounceY = 1
	case 16:
		f.bounceY = 0
		f.bouncei = 0
		playSnd(snd.fruitBounce)
	}

	if f.slowTimer++; f.slowTimer == 2 {
		f.slowTimer = 0
		f.pos = f.pos.Add(f.vel)

		f.nearest.X = (f.pos.X + Tile.X/2) / Tile.X
		f.nearest.Y = (f.pos.Y + Tile.Y/2) / Tile.Y

		if f.pos.X%Tile.X == 0 && f.pos.Y%Tile.Y == 0 {
			if len(f.path) > 0 {
				f.path = f.path[1:]
				f.FollowNextPath()
			} else {
				f.pos.X = f.nearest.X * Tile.X
				f.pos.Y = f.nearest.Y * Tile.Y
				f.active = false
				game.fruitTimer = 0
			}
		}
	}
}

func (f *Fruit) FollowNextPath() {
	if !f.hasPath {
		return
	}

	if len(f.path) == 0 {
		return
	}

	switch f.path[0] {
	case 'L':
		f.vel = sdl.Point{int32(-f.speed), 0}
	case 'R':
		f.vel = sdl.Point{int32(f.speed), 0}
	case 'U':
		f.vel = sdl.Point{0, int32(-f.speed)}
	case 'D':
		f.vel = sdl.Point{0, int32(f.speed)}
	}
}

func (f *Fruit) Draw() {
	if game.mode == 3 || !f.active {
		return
	}

	f.images[f.typ].Draw(sdl.Point{
		f.pos.X - game.pixelPos.X,
		f.pos.Y - game.pixelPos.Y - int32(f.bounceY),
	})
}
