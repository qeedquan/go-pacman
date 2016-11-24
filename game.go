package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
)

type HiScore struct {
	name  string
	score int
}

type HiScoreSlice []HiScore

func (p HiScoreSlice) Len() int           { return len(p) }
func (p HiScoreSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p HiScoreSlice) Less(i, j int) bool { return p[i].score < p[j].score }

type Game struct {
	background *Image
	life       *Image
	gameOver   *Image
	ready      *Image
	logo       *Image
	hiscore    *Image
	digits     []*Image

	mode  int
	level int
	score int
	lives int

	modeTimer int

	fruitTimer      int
	fruitScoreTimer int

	ghostTimer int
	ghostValue int

	pixelPos       sdl.Point
	nearestTilePos sdl.Point
	pixelOffset    sdl.Point
}

func newGame() *Game {
	var digits [10]*Image
	for i := range digits {
		digits[i] = loadImage(fmt.Sprintf("text/%d.gif", i), colorKey(sdlcolor.White))
	}

	g := &Game{
		background: loadImage("backgrounds/1.gif"),
		life:       loadImage("text/life.gif"),
		gameOver:   loadImage("text/gameover.gif"),
		ready:      loadImage("text/ready.gif"),
		logo:       loadImage("text/logo.gif"),
		digits:     digits[:],
		hiscore:    makeTexture(HSSize),
		lives:      3,
	}
	g.SetMode(3)
	g.MakeHiScoreList()
	return g
}

func (g *Game) Start() {
	g.level = 1
	g.score = 0
	g.lives = 3
	g.SetMode(4)
	level.Load(g.level)
}

func (g *Game) SetMode(mode int) {
	g.mode = mode
	g.modeTimer = 0
}

func defaultHiScores() []HiScore {
	return []HiScore{
		{"David", 100000},
		{"Andy", 80000},
		{"Count Pacula", 60000},
		{"Cleopacra", 40000},
		{"Brett Favre", 20000},
		{"Sergei Pachmanioff", 10000},
	}
}

func (g *Game) GetHiScores() (list []HiScore) {
	var err error

	defer func() {
		if err != nil {
			list = defaultHiScores()
		}
	}()

	filename := filepath.Join(pref, "hiscore.txt")
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		var p HiScore
		line := s.Text()
		n, err := fmt.Sscan(line, &p.score, &p.name)
		if n != 2 || err != nil {
			continue
		}

		if p.score > MaxScore {
			p.score = MaxScore
		}

		s := ""
		i := 0
		for _, r := range p.name {
			s += string(r)
			if i++; i == 22 {
				break
			}
		}
		p.name = s

		list = append(list, p)
		if len(list) == 6 {
			sort.Stable(HiScoreSlice(list))
			list = list[:6]
		}
	}

	for len(list) < 6 {
		list = append(list, HiScore{})
	}

	return
}

func (g *Game) WriteHiScores(hs []HiScore) {
	log.SetPrefix("score: ")
	filename := filepath.Join(pref, "hiscore.txt")
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	for _, p := range hs {
		fmt.Fprint(f, "%d %s\n", p.score, p.name)
	}
}

func (g *Game) UpdateHiScores(score int) {
	hs := g.GetHiScores()
	hs = append(hs, HiScore{userName, score})
	sort.Stable(HiScoreSlice(hs))
	g.WriteHiScores(hs)
}

func (g *Game) MakeHiScoreList() {
	g.hiscore.Bind()
	screen.SetDrawColor(sdlcolor.Black)
	screen.Clear()
	g.hiscore.tex.SetAlphaMod(HSAlpha)

	blitText(font, 0, 0, sdl.Color{255, 255, 255, 0}, fmt.Sprintf("%18s HIGH SCORES", " "))
	hs := g.GetHiScores()
	y := 0
	for _, h := range hs {
		y += HSLineHeight
		score := fmt.Sprint(h.score)
		name := h.name
		text := fmt.Sprintf("%*s%s%*s%s", 22-len(name), " ", name, 9-len(score), " ", score)
		blitText(font, 0, y, sdlcolor.White, text)
	}

	g.hiscore.Unbind()
}

func (g *Game) DrawMidGameHiScores() {
	g.MakeHiScoreList()
}

func (g *Game) DrawScore() {
	g.DrawNumber(g.score, sdl.Point{Score.X, Scr.Y - Score.Y})

	for i := 0; i < g.lives; i++ {
		g.life.Draw(sdl.Point{34 + int32(i)*10 + 16, Scr.Y - 18})
	}

	fruit.images[fruit.typ].Draw(sdl.Point{20, Scr.Y - 28})

	switch g.mode {
	case 3:
		p := Scr.Sub(g.gameOver.size)
		p.X, p.Y = p.X/2, p.Y/2
		g.gameOver.Draw(p)
	case 4:
		g.ready.Draw(sdl.Point{Scr.X/2 - 20, Scr.Y/2 + 12})
	}

	g.DrawNumber(g.level, sdl.Point{0, Scr.Y - 20})
}

func (g *Game) DrawNumber(n int, p sdl.Point) {
	s := fmt.Sprint(n)
	for i, r := range s {
		g.digits[r-'0'].Draw(sdl.Point{p.X + int32(i*ScoreColumn), p.Y})
	}
}

func (g *Game) SmartMoveScreen() {
	x := player.pos.X - Grid.X/2*Tile.X
	y := player.pos.Y - Grid.Y/2*Tile.Y

	x = int32(clamp(int(x), 0, int(level.size.X*Tile.X-Scr.X)))
	y = int32(clamp(int(y), 0, int(level.size.Y*Tile.Y-Scr.Y)))
	g.MoveScreen(sdl.Point{x, y})
}

func (g *Game) MoveScreen(p sdl.Point) {
	g.pixelPos = p
	g.nearestTilePos = sdl.Point{p.X / Tile.X, p.Y / Tile.Y}
	g.pixelOffset = sdl.Point{
		p.X - g.nearestTilePos.X*Tile.X,
		p.Y - g.nearestTilePos.Y*Tile.Y,
	}
}

func (g *Game) NextLevel() {
	g.level++
	g.SetMode(4)
	level.Load(g.level)

	player.vel = sdl.Point{}
	player.anim = player.animS
}

func (g *Game) AddScore(amount int) {
	extraLifeSet := []int{25000, 50000, 100000, 150000}
	for _, n := range extraLifeSet {
		if g.score < n && g.score+amount >= n {
			playSnd(snd.extraLife)
			g.lives++
		}
	}
	g.score += amount
}
