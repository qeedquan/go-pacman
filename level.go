package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
)

type Level struct {
	size            sdl.Point
	edgeShadowColor sdl.Color
	edgeLightColor  sdl.Color
	fillColor       sdl.Color
	pelletColor     sdl.Color

	grid                  []int
	fruitType             int
	pellets               int
	powerPelletBlinkTimer int
}

func newLevel() *Level {
	return &Level{
		edgeLightColor:  sdl.Color{255, 255, 0, 255},
		edgeShadowColor: sdl.Color{255, 150, 0, 255},
		fillColor:       sdl.Color{0, 255, 255, 255},
		pelletColor:     sdl.Color{255, 255, 255, 255},
	}
}

func (l *Level) reset() {
	*l = Level{}
}

func (l *Level) Load(level int) {
	log.SetPrefix("level: ")
	filename := fmt.Sprintf("%s/levels/%d.txt", assets, level)
	filename = filepath.Clean(filename)
	f, err := os.Open(filename)
	ck(err)
	defer f.Close()

	l.reset()

	levelData := false
	y := 0
	nline := 1
	s := bufio.NewScanner(f)
	for ; s.Scan(); nline++ {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		useLine := true
		if fields[0] == "#" && len(fields) >= 2 {
			useLine = false
			switch fields[1] {
			case "lvlwidth":
				l.size.X = int32(fieldInt(fields))
			case "lvlheight":
				l.size.Y = int32(fieldInt(fields))
			case "edgecolor":
				l.edgeLightColor = fieldRGB(fields)
				l.edgeShadowColor = l.edgeLightColor
			case "edgelightcolor":
				l.edgeLightColor = fieldRGB(fields)
			case "edgeshadowcolor":
				l.edgeShadowColor = fieldRGB(fields)
			case "fillcolor":
				l.fillColor = fieldRGB(fields)
			case "pelletcolor":
				l.pelletColor = fieldRGB(fields)
			case "fruittype":
				l.fruitType = fieldInt(fields)
			case "startleveldata":
				y = 0
				levelData = true
			case "endleveldata":
				levelData = false
			}
		}

		if !useLine {
			continue
		}

		if levelData {
			if len(fields) != int(l.size.X) {
				log.Fatalf("invalid level data, got %d width for line %d, want %d\n", level, nline, l.size.X)
			}

			if len(l.grid) < int(l.size.X*l.size.Y) {
				l.grid = make([]int, l.size.X*l.size.Y)
			}

			for x := int32(0); x < l.size.X; x++ {
				id, _ := strconv.Atoi(fields[x])
				i := sdl.Point{x, int32(y)}
				l.SetTile(i, id)

				switch {
				case id == 4:
					player.home = sdl.Point{int32(x) * Tile.X, int32(y) * Tile.Y}
					l.SetTile(i, 0)

				case 10 <= id && id <= 13:
					ghosts[id-10].home = sdl.Point{int32(x) * Tile.X, int32(y) * Tile.Y}
					l.SetTile(i, 0)

				case id == 2:
					l.pellets++
				}
			}
			y++
		}
	}

	if len(l.grid) < int(l.size.X*l.size.Y) {
		log.Fatalf("grid size too small, got %d size; expected size %dx%d (%d)",
			len(l.grid), l.size.X, l.size.Y, l.size.X*l.size.Y)
	}

	tiles.Load()
	path.Resize(l.size)
	for y := int32(0); y < path.size.Y; y++ {
		for x := int32(0); x < path.size.X; x++ {
			i := sdl.Point{x, y}
			if l.IsWall(i) {
				path.SetType(i, 1)
			} else {
				path.SetType(i, 0)
			}
		}
	}

	l.Restart()
}

func (l *Level) Restart() {
	for i := 0; i < 4; i++ {
		g := ghosts[i]
		g.pos = g.home
		g.vel = sdl.Point{}
		g.state = 1
		g.speed = 1
		g.Move()
		g.RandPelletSpot()
	}

	player.pos = player.home
	player.vel = sdl.Point{}
	player.frame = 3
	player.anim = player.animS

	fruit.active = false
	game.fruitTimer = 0
}

func (l *Level) GhostBoxPos() sdl.Point {
	for y := int32(0); y < l.size.Y; y++ {
		for x := int32(0); x < l.size.X; x++ {
			i := sdl.Point{x, y}
			if l.Tile(i) == tiles.id["ghost-door"] {
				return i
			}
		}
	}
	return sdl.Point{-1, -1}
}

func (l *Level) SetTile(pos sdl.Point, v int) {
	l.grid[pos.Y*l.size.X+pos.X] = v
}

func (l *Level) Tile(pos sdl.Point) int {
	if l.outOfBounds(pos) {
		return 0
	}
	return l.grid[pos.Y*l.size.X+pos.X]
}

func (l *Level) outOfBounds(pos sdl.Point) bool {
	if pos.X < 0 || pos.X >= l.size.X {
		return true
	}
	if pos.Y < 0 || pos.Y >= l.size.Y {
		return true
	}
	return false
}

func (l *Level) IsWall(pos sdl.Point) bool {
	if l.outOfBounds(pos) {
		return true
	}

	id := l.Tile(pos)
	return 100 <= id && id <= 199
}

func (l *Level) CheckIfHitWall(possiblePlayer, pos sdl.Point) bool {
	for y := pos.Y - 1; y < pos.Y+2; y++ {
		for x := pos.X - 1; x < pos.X+2; x++ {
			if !inTile(possiblePlayer, sdl.Point{x, y}) {
				continue
			}

			if l.IsWall(sdl.Point{x, y}) {
				return true
			}
		}
	}
	return false
}

func (l *Level) CheckIfHitSomething(playerPos sdl.Point, pos sdl.Point) {
	for y := pos.Y - 1; y < pos.Y+2; y++ {
		for x := pos.X - 1; x < pos.X+2; x++ {
			i := sdl.Point{x, y}
			if !inTile(playerPos, i) {
				continue
			}

			id := level.Tile(i)
			switch id {
			case tiles.id["pellet"]:
				l.SetTile(i, 0)
				playSnd(snd.pellet[player.pelletSnd])
				player.pelletSnd = 1 - player.pelletSnd
				l.pellets--
				game.AddScore(10)

				if l.pellets == 0 {
					game.SetMode(6)
				}

			case tiles.id["pellet-power"]:
				l.SetTile(i, 0)
				playSnd(snd.powerPellet)
				game.AddScore(100)
				game.ghostValue = 200

				game.ghostTimer = 360
				for i := 0; i < 4; i++ {
					if ghosts[i].state == 1 {
						ghosts[i].state = 2
					}
				}

			case tiles.id["door-h"]:
				for dx := int32(0); dx < l.size.X; dx++ {
					if dx != x {
						if l.Tile(sdl.Point{dx, y}) == tiles.id["door-h"] {
							player.pos.X = dx * Tile.X
							if player.vel.X > 0 {
								player.pos.X += Tile.X
							} else {
								player.pos.X -= Tile.X
							}
						}
					}
				}

			case tiles.id["door-v"]:
				for dy := int32(0); dy < l.size.Y; dy++ {
					if dy != y {
						if l.Tile(sdl.Point{x, dy}) == tiles.id["door-v"] {
							player.pos.Y = dy * Tile.Y
							if player.vel.Y > 0 {
								player.pos.Y += Tile.Y
							} else {
								player.pos.Y -= Tile.Y
							}
						}
					}
				}
			}
		}
	}
}

func (l *Level) CheckIfHit(playerPos sdl.Point, pos sdl.Point, cushion int) bool {
	d := playerPos.Sub(pos)
	if abs(int(d.X)) < cushion && abs(int(d.Y)) < cushion {
		return true
	}
	return false
}

func (l *Level) Draw() {
	l.powerPelletBlinkTimer = (l.powerPelletBlinkTimer + 1) % 60

	for y := int32(-1); y < Grid.Y+1; y++ {
		for x := int32(-1); x < Grid.X+1; x++ {
			ax := game.nearestTilePos.X + x
			ay := game.nearestTilePos.Y + y
			id := l.Tile(sdl.Point{ax, ay})

			if id == 0 || id == tiles.id["door-h"] || id == tiles.id["door-v"] {
				continue
			}

			switch id {
			case tiles.id["pellet-power"]:
				if l.powerPelletBlinkTimer < 30 {
					tiles.image[id].Draw(sdl.Point{
						x*Tile.X - game.pixelOffset.X,
						y*Tile.Y - game.pixelOffset.Y,
					})
				}

			case tiles.id["showlogo"]:
				game.logo.Draw(sdl.Point{
					x*Tile.X - game.pixelOffset.X,
					y*Tile.Y - game.pixelOffset.Y,
				})

			case tiles.id["hiscores"]:
				game.hiscore.Draw(sdl.Point{
					x*Tile.X - game.pixelOffset.X,
					y*Tile.Y - game.pixelOffset.Y,
				})

			default:
				tiles.image[id].Draw(sdl.Point{
					x*Tile.X - game.pixelOffset.X,
					y*Tile.Y - game.pixelOffset.Y,
				})

			}
		}
	}
}

func (l *Level) PathPairPos() (start, end sdl.Point) {
	var door []sdl.Point

	start = sdl.Point{-1, -1}
	end = sdl.Point{-1, -1}

	for y := int32(0); y < l.size.Y; y++ {
		for x := int32(0); x < l.size.X; x++ {
			i := sdl.Point{x, y}
			switch l.Tile(i) {
			case tiles.id["door-h"], tiles.id["door-v"]:
				door = append(door, i)
			}
		}
	}

	if len(door) == 0 {
		return
	}

	chosenDoor := door[rand.Intn(len(door))]
	if l.Tile(chosenDoor) == tiles.id["door-h"] {
		for x := int32(0); x < l.size.X; x++ {
			if x != chosenDoor.X {
				if l.Tile(sdl.Point{x, chosenDoor.Y}) == tiles.id["door-h"] {
					return chosenDoor, sdl.Point{x, chosenDoor.Y}
				}
			}
		}
	} else {
		for y := int32(0); y < l.size.Y; y++ {
			if y != chosenDoor.Y {
				if l.Tile(sdl.Point{chosenDoor.X, y}) == tiles.id["door-v"] {
					return chosenDoor, sdl.Point{chosenDoor.X, y}
				}
			}
		}
	}

	return
}

func fieldInt(fields []string) int {
	if len(fields) < 3 {
		return 0
	}
	n, _ := strconv.Atoi(fields[2])
	return n
}

func fieldRGB(fields []string) sdl.Color {
	if len(fields) < 5 {
		return sdlcolor.Transparent
	}

	r, _ := strconv.Atoi(fields[2])
	g, _ := strconv.Atoi(fields[3])
	b, _ := strconv.Atoi(fields[4])
	return sdl.Color{uint8(r), uint8(g), uint8(b), 255}
}

func inTile(p, s sdl.Point) bool {
	if !(p.X-s.X*Tile.X < Tile.X) {
		return false
	}

	if !(p.X-s.X*Tile.X > -Tile.X) {
		return false
	}

	if !(p.Y-s.Y*Tile.Y < Tile.Y) {
		return false
	}

	if !(p.Y-s.Y*Tile.Y > -Tile.Y) {
		return false
	}

	return true
}
