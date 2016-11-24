package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
	"github.com/qeedquan/go-media/sdl/sdlmixer"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

type Display struct {
	*sdl.Window
	*sdl.Renderer
}

const (
	MaxScore = 1e9

	HSFontSize   = 14
	HSAlpha      = 200
	HSLineHeight = 16
)

var (
	Tile = sdl.Point{24, 24}
	Grid = sdl.Point{21, 23}
	Scr  = sdl.Point{Grid.X * Tile.X, Grid.Y * Tile.Y}

	Score       = sdl.Point{50, 34}
	ScoreColumn = 13

	HSPos  = sdl.Point{48, 384}
	HSSize = sdl.Point{408, 120}
)

var (
	assets     string
	pref       string
	fullscreen bool
	sfx        bool
	userName   string
	infLives   bool
)

var (
	screen  *Display
	tiles   *Tiles
	game    *Game
	level   *Level
	path    *Path
	player  *Pacman
	ghosts  [6]*Ghost
	fruit   *Fruit
	snd     *Snd
	font    *sdlttf.Font
	texture *sdl.Texture
	surface *sdl.Surface
)

func main() {
	runtime.LockOSThread()
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)
	parseFlags()
	initSDL()
	load()
	play()
}

func newDisplay(w, h int, wflag sdl.WindowFlags) (*Display, error) {
	window, renderer, err := sdl.CreateWindowAndRenderer(w, h, wflag)
	if err != nil {
		return nil, err
	}
	return &Display{window, renderer}, nil
}

func parseFlags() {
	assets = filepath.Join(sdl.GetBasePath(), "assets")
	pref = sdl.GetPrefPath("", "pacman")
	flag.StringVar(&assets, "assets", assets, "data directory")
	flag.StringVar(&pref, "pref", pref, "data directory")
	flag.BoolVar(&fullscreen, "fullscreen", false, "fullscreen")
	flag.BoolVar(&sfx, "sfx", true, "enable sfx")
	flag.StringVar(&userName, "user", "User", "user name")
	flag.BoolVar(&infLives, "inflives", false, "infinite lives")
	flag.Parse()
}

func initSDL() {
	log.SetPrefix("sdl: ")
	err := sdl.Init(sdl.INIT_EVERYTHING &^ sdl.INIT_AUDIO)
	ck(err)

	err = sdl.InitSubSystem(sdl.INIT_AUDIO)
	ek(err)

	err = sdlmixer.OpenAudio(44100, sdl.AUDIO_S16, 2, 8192)
	ek(err)

	err = sdlttf.Init()
	ck(err)

	w, h := int(Scr.X), int(Scr.Y)
	wflag := sdl.WINDOW_RESIZABLE
	if fullscreen {
		wflag |= sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	screen, err = newDisplay(w, h, wflag)
	ck(err)

	texture, err = screen.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, w, h)
	ck(err)

	surface, err = sdl.CreateRGBSurfaceWithFormat(sdl.SWSURFACE, w, h, 32, sdl.PIXELFORMAT_ABGR8888)
	ck(err)

	screen.SetTitle("Pacman")
	screen.SetLogicalSize(int(Scr.X), int(Scr.Y))
	screen.SetDrawColor(sdlcolor.Black)
	screen.Clear()
	screen.Present()

	sdl.ShowCursor(0)
}

func load() {
	font = loadFont("VeraMoBd.ttf", HSFontSize)
	game = newGame()
	tiles = newTiles()
	path = newPath()
	player = newPacman()
	for i := range ghosts {
		ghosts[i] = newGhost(i)
	}
	fruit = newFruit()
	snd = newSnd()

	level = newLevel()
	level.Load(game.level)
}

func play() {
	for {
		checkClose()
		update()
		draw()
		sdl.Delay(1000 / 60)
	}
}

func checkClose() {
	for {
		ev := sdl.PollEvent()
		if ev == nil {
			break
		}
		switch ev := ev.(type) {
		case sdl.QuitEvent:
			os.Exit(0)
		case sdl.KeyDownEvent:
			switch ev.Sym {
			case sdl.K_ESCAPE:
				os.Exit(0)
			}
		}
	}
}

func checkInputs() {
	k := sdl.GetKeyboardState()
	switch game.mode {
	case 1:
		switch {
		case k[sdl.SCANCODE_RIGHT] != 0:
			p := sdl.Point{player.pos.X + int32(player.speed), player.pos.Y}
			q := player.nearest
			if !(player.vel.X == int32(player.speed) && player.vel.Y == 0) && !level.CheckIfHitWall(p, q) {
				player.vel.X = int32(player.speed)
				player.vel.Y = 0
			}

		case k[sdl.SCANCODE_LEFT] != 0:
			p := sdl.Point{player.pos.X - int32(player.speed), player.pos.Y}
			q := player.nearest
			if !(player.vel.X == int32(-player.speed) && player.vel.Y == 0) && !level.CheckIfHitWall(p, q) {
				player.vel.X = int32(-player.speed)
				player.vel.Y = 0
			}

		case k[sdl.SCANCODE_DOWN] != 0:
			p := sdl.Point{player.pos.X, player.pos.Y + int32(player.speed)}
			q := player.nearest
			if !(player.vel.X == 0 && player.vel.Y == int32(player.speed)) && !level.CheckIfHitWall(p, q) {
				player.vel.X = 0
				player.vel.Y = int32(player.speed)
			}

		case k[sdl.SCANCODE_UP] != 0:
			p := sdl.Point{player.pos.X, player.pos.Y - int32(player.speed)}
			q := player.nearest
			if !(player.vel.X == 0 && player.vel.Y == int32(-player.speed)) && !level.CheckIfHitWall(p, q) {
				player.vel.X = 0
				player.vel.Y = int32(-player.speed)
			}
		}

	case 3:
		switch {
		case k[sdl.SCANCODE_RETURN] != 0 || k[sdl.SCANCODE_KP_ENTER] != 0:
			game.Start()
		}
	}
}

var (
	oldEdgeLightColor  sdl.Color
	oldEdgeShadowColor sdl.Color
	oldFillColor       sdl.Color
)

func update() {
	switch game.mode {
	case 1: // normal gameplay
		checkInputs()
		game.modeTimer++
		player.Move()
		for i := 0; i < 4; i++ {
			ghosts[i].Move()
		}
		fruit.Move()

	case 2: // waiting after getting hit by ghost
		if game.modeTimer++; game.modeTimer == 90 {
			level.Restart()
			if !infLives {
				game.lives--
			}
			if game.lives < 0 {
				game.UpdateHiScores(game.score)
				game.SetMode(3)
				game.DrawMidGameHiScores()
			} else {
				game.SetMode(4)
			}
		}

	case 3: // game over
		checkInputs()

	case 4: // waiting to start
		if game.modeTimer++; game.modeTimer == 90 {
			game.SetMode(1)
			player.vel.X = int32(player.speed)
		}

	case 5: // brief pause after munching vurnerable ghost
		if game.modeTimer++; game.modeTimer == 30 {
			game.SetMode(1)
		}

	case 6: // pause after eating all the pellets
		if game.modeTimer++; game.modeTimer == 60 {
			game.SetMode(7)
			oldEdgeLightColor = level.edgeLightColor
			oldEdgeShadowColor = level.edgeShadowColor
			oldFillColor = level.fillColor
		}

	case 7: // flashing maze after finishing a level
		whiteSet := []int{10, 30, 50, 70}
		normalSet := []int{20, 40, 60, 80}

		game.modeTimer++
		switch {
		case inIntSet(whiteSet, game.modeTimer):
			level.edgeLightColor = sdl.Color{255, 255, 254, 255}
			level.edgeShadowColor = sdl.Color{255, 255, 254, 255}
			level.fillColor = sdl.Color{0, 0, 0, 255}
			tiles.Load()

		case inIntSet(normalSet, game.modeTimer):
			level.edgeLightColor = oldEdgeLightColor
			level.edgeShadowColor = oldEdgeShadowColor
			level.fillColor = oldFillColor
			tiles.Load()

		case game.modeTimer == 150:
			game.SetMode(8)
		}

	case 8: // blank screen before changing levels
		if game.modeTimer++; game.modeTimer == 10 {
			game.NextLevel()
		}
	}

	game.SmartMoveScreen()
}

func draw() {
	screen.SetDrawColor(sdlcolor.Black)
	screen.Clear()

	game.background.Draw(sdl.Point{})

	if game.mode != 8 {
		level.Draw()
		for i := 0; i < 4; i++ {
			ghosts[i].Draw()
		}
		fruit.Draw()
		player.Draw()

		if game.mode == 3 {
			game.hiscore.Draw(HSPos)
		}
	}

	if game.mode == 5 {
		game.DrawNumber(game.ghostValue/2, sdl.Point{
			player.pos.X - game.pixelPos.X - 4,
			player.pos.Y - game.pixelPos.Y + 6,
		})
	}

	game.DrawScore()
	screen.Present()
}

func inIntSet(l []int, v int) bool {
	for i := range l {
		if l[i] == v {
			return true
		}
	}
	return false
}

func blitText(font *sdlttf.Font, x, y int, c sdl.Color, text string) {
	r, err := font.RenderUTF8BlendedEx(surface, text, c)
	ck(err)

	p, err := texture.Lock(nil)
	ck(err)

	err = surface.Lock()
	ck(err)

	s := surface.Pixels()
	for i := 0; i < len(p); i += 4 {
		p[i] = s[i+2]
		p[i+1] = s[i]
		p[i+2] = s[i+1]
		p[i+3] = s[i+3]
	}

	surface.Unlock()
	texture.Unlock()

	texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	screen.Copy(texture, &sdl.Rect{0, 0, r.W, r.H}, &sdl.Rect{int32(x), int32(y), r.W, r.H})
}
