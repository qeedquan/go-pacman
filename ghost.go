package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/qeedquan/go-media/sdl"
)

type Ghost struct {
	Entity
	path      string
	hasPath   bool
	id        int
	frame     int
	animDelay int
	anim      []*Image
}

var ghostColors = []sdl.Color{
	{255, 0, 0, 255},
	{255, 128, 255, 255},
	{128, 255, 255, 255},
	{255, 128, 0, 255},
	{50, 50, 255, 255},
	{255, 255, 255, 255},
}

func newGhost(id int) *Ghost {
	var anim []*Image

	log.SetPrefix("image: ")
	anim = make([]*Image, 8)
	for i := 1; i < 7; i++ {
		img := loadRGBA(fmt.Sprintf("sprite/ghost %d.gif", i))
		for y := 0; y < int(Tile.Y); y++ {
			for x := 0; x < int(Tile.X); x++ {
				if (img.RGBAAt(x, y) == color.RGBA{255, 0, 0, 255}) {
					img.Set(x, y, ghostColors[id])
				}
			}
		}

		anim[i] = makeImage(img)
	}

	return &Ghost{
		Entity: Entity{
			state: 1,
			speed: 1,
		},
		id:    id,
		anim:  anim,
		frame: 1,
	}
}

func (g *Ghost) FollowNextPath() {
	if !g.hasPath {
		return
	}

	if len(g.path) > 0 {
		switch g.path[0] {
		case 'L':
			g.vel = sdl.Point{int32(-g.speed), 0}
		case 'R':
			g.vel = sdl.Point{int32(g.speed), 0}
		case 'U':
			g.vel = sdl.Point{0, int32(-g.speed)}
		case 'D':
			g.vel = sdl.Point{0, int32(g.speed)}
		}
	} else {
		if g.state != 3 {
			g.path, g.hasPath = path.Find(g.nearest, player.nearest)
			g.FollowNextPath()
		} else {
			g.state = 1
			g.speed /= 4
			g.RandPelletSpot()
		}
	}
}

func (g *Ghost) RandPelletSpot() {
	log.SetPrefix("ghost: ")

	var pelletSpots []sdl.Point
	for y := int32(0); y < level.size.Y-2; y++ {
		for x := int32(0); x < level.size.X-2; x++ {
			i := sdl.Point{x, y}
			if level.Tile(i) == tiles.id["pellet"] {
				pelletSpots = append(pelletSpots, i)
			}
		}
	}

	if len(pelletSpots) == 0 {
		log.Fatal("no pellet spots left for spawning")
	}

	r := pelletSpots[rand.Intn(len(pelletSpots))]
	g.path, g.hasPath = path.Find(g.nearest, r)
	g.FollowNextPath()
}

func (g *Ghost) Move() {
	g.pos = g.pos.Add(g.vel)

	g.nearest = sdl.Point{
		(g.pos.X + Tile.X/2) / Tile.X,
		(g.pos.Y + Tile.Y/2) / Tile.Y,
	}

	if g.pos.X%Tile.X == 0 && g.pos.Y%Tile.Y == 0 {
		if len(g.path) > 0 {
			g.path = g.path[1:]
			g.FollowNextPath()
		} else {
			g.pos = sdl.Point{g.nearest.X * Tile.X, g.nearest.Y * Tile.Y}
			g.path, g.hasPath = path.Find(g.nearest, player.nearest)
			g.FollowNextPath()
		}
	}
}

func (g *Ghost) Draw() {
	if game.mode == 3 {
		return
	}

	m := g.anim[g.frame]

	// draw ghost eyes
	m.Bind()
	for _, y := range []int{6, 12, 1} {
		for _, x := range []int{5, 6, 8, 9} {
			screen.SetDrawColor(sdl.Color{0xf8, 0xf8, 0xf8, 0xff})
			screen.DrawPoint(x, y)
			screen.DrawPoint(x+9, y)
		}
	}
	m.Unbind()

	switch g.state {
	case 1:
		m.Draw(g.pos.Sub(game.pixelPos))

	case 2:
		if game.ghostTimer > 100 {
			ghosts[4].anim[g.frame].Draw(g.pos.Sub(game.pixelPos))
		} else {
			t := game.ghostTimer / 10
			if t == 1 || t == 3 || t == 5 || t == 7 || t == 9 {
				ghosts[5].anim[g.frame].Draw(g.pos.Sub(game.pixelPos))
			} else {
				ghosts[5].anim[g.frame].Draw(g.pos.Sub(game.pixelPos))
			}
		}

	case 3:
		m := tiles.image[tiles.id["glasses"]]
		m.Draw(g.pos.Sub(game.pixelPos))
	}

	// don't animate ghost if the level is complete
	if game.mode == 6 || game.mode == 7 {
		return
	}

	if g.animDelay++; g.animDelay == 2 {
		if g.frame++; g.frame == 7 {
			g.frame = 1
		}

		g.animDelay = 0
	}
}
