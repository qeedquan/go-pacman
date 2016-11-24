package main

import (
	"fmt"

	"github.com/qeedquan/go-media/sdl"
)

type Pacman struct {
	Entity
	animL     []*Image
	animR     []*Image
	animS     []*Image
	animU     []*Image
	animD     []*Image
	anim      []*Image
	frame     int
	pelletSnd int
}

func newPacman() *Pacman {
	l := make([]*Image, 10)
	r := make([]*Image, 10)
	s := make([]*Image, 10)
	u := make([]*Image, 10)
	d := make([]*Image, 10)
	for i := 1; i < 9; i++ {
		l[i] = loadImage(fmt.Sprintf("sprite/pacman-l %d.gif", i))
		r[i] = loadImage(fmt.Sprintf("sprite/pacman-r %d.gif", i))
		u[i] = loadImage(fmt.Sprintf("sprite/pacman-u %d.gif", i))
		d[i] = loadImage(fmt.Sprintf("sprite/pacman-d %d.gif", i))
		s[i] = loadImage("sprite/pacman.gif")
	}

	return &Pacman{
		Entity: Entity{
			speed: 3,
		},
		animL: l,
		animR: r,
		animS: s,
		animU: u,
		animD: d,
	}
}

func (p *Pacman) Move() {
	p.nearest.X = (p.pos.X + Tile.X/2) / Tile.X
	p.nearest.Y = (p.pos.Y + Tile.Y/2) / Tile.Y

	if !level.CheckIfHitWall(p.pos.Add(p.vel), p.nearest) {
		p.pos = p.pos.Add(p.vel)
		level.CheckIfHitSomething(p.pos, p.nearest)

		for i := 0; i < 4; i++ {
			g := ghosts[i]
			if level.CheckIfHit(p.pos, g.pos, int(Tile.X/2)) {
				switch g.state {
				case 1:
					game.SetMode(2)
				case 2:
					game.AddScore(game.ghostValue)
					game.ghostValue *= 2
					playSnd(snd.eatgh)

					g.state = 3
					g.speed *= 4
					g.pos = sdl.Point{g.nearest.X * Tile.X, g.nearest.Y * Tile.Y}

					boxPos := level.GhostBoxPos()
					boxPos.Y++
					g.path, g.hasPath = path.Find(g.nearest, boxPos)
					g.FollowNextPath()

					game.SetMode(5)
				}
			}
		}

		if fruit.active {
			if level.CheckIfHit(p.pos, fruit.pos, int(Tile.X/2)) {
				game.AddScore(2500)
				fruit.active = false
				game.fruitTimer = 0
				game.fruitScoreTimer = 120
				playSnd(snd.eatFruit)
			}
		}
	} else {
		p.vel = sdl.Point{}
	}

	if game.ghostTimer > 0 {
		if game.ghostTimer--; game.ghostTimer == 0 {
			for _, g := range ghosts {
				if g.state == 2 {
					g.state = 1
				}
			}
		}
	}

	if game.fruitTimer++; game.fruitTimer == 500 {
		pathEntrance, pathExit := level.PathPairPos()
		if !fruit.active && (pathEntrance != sdl.Point{-1, -1}) && (pathExit != sdl.Point{-1, -1}) {
			fruit.active = true
			fruit.nearest = pathEntrance
			fruit.pos = sdl.Point{
				fruit.nearest.X * Tile.X,
				fruit.nearest.Y * Tile.Y,
			}
			fruit.path, fruit.hasPath = path.Find(fruit.nearest, pathExit)
			fruit.FollowNextPath()
		}
	}

	if game.fruitScoreTimer > 0 {
		game.fruitScoreTimer--
	}
}

func (p *Pacman) Draw() {
	if game.mode == 3 {
		return
	}

	switch {
	case p.vel.X > 0:
		p.anim = p.animR
	case p.vel.X < 0:
		p.anim = p.animL
	case p.vel.Y > 0:
		p.anim = p.animD
	case p.vel.Y < 0:
		p.anim = p.animU
	}

	p.anim[p.frame].Draw(p.pos.Sub(game.pixelPos))

	if game.mode == 1 {
		if p.vel.X != 0 || p.vel.Y != 0 {
			p.frame++
		}

		if p.frame == 9 {
			p.frame = 1
		}
	}
}
