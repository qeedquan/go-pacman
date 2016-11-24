package main

import (
	"bufio"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Tiles struct {
	name  map[int]string
	id    map[string]int
	image map[int]*Image
}

func newTiles() *Tiles {
	return &Tiles{
		name:  make(map[int]string),
		id:    make(map[string]int),
		image: make(map[int]*Image),
	}
}

func (t *Tiles) reset() {
	for _, m := range t.image {
		m.Free()
	}

	for k := range t.name {
		delete(t.name, k)
	}
	for k := range t.id {
		delete(t.id, k)
	}
	for k := range t.image {
		delete(t.image, k)
	}
}

func (t *Tiles) Load() {
	var noGifTiles = []int{23}
	var (
		edgeLightColor  = color.RGBA{0xff, 0xce, 0xff, 0xff}
		fillColor       = color.RGBA{0x84, 0x00, 0x84, 0xff}
		edgeShadowColor = color.RGBA{0xff, 0x00, 0xff, 0xff}
		pelletColor     = color.RGBA{0x80, 0x00, 0x80, 0xff}
	)

	log.SetPrefix("tiles: ")
	filename := filepath.Join(assets, "crossref.txt")
	f, err := os.Open(filename)
	ck(err)
	defer f.Close()

	t.reset()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "#" || fields[0] == "'" {
			continue
		}

		if len(fields) < 2 {
			continue
		}

		id, _ := strconv.Atoi(fields[0])
		name := fields[1]

		noGif := false
		for _, tid := range noGifTiles {
			if tid == id {
				noGif = true
				break
			}
		}

		var img *image.RGBA
		if !noGif {
			img = loadRGBA("tiles/" + name + ".gif")
		} else {
			img = image.NewRGBA(image.Rect(0, 0, int(Tile.X), int(Tile.Y)))
		}

		for y := 0; y < int(Tile.Y); y++ {
			for x := 0; x < int(Tile.X); x++ {
				switch img.RGBAAt(x, y) {
				case edgeLightColor:
					img.Set(x, y, level.edgeLightColor)
				case fillColor:
					img.Set(x, y, level.fillColor)
				case edgeShadowColor:
					img.Set(x, y, level.edgeShadowColor)
				case pelletColor:
					img.Set(x, y, level.pelletColor)
				}
			}
		}

		t.name[id] = name
		t.id[name] = id
		t.image[id] = makeImage(img)
	}

	t.verify()
}

func (t *Tiles) verify() {
	log.SetPrefix("tiles: ")
	names := []string{"pellet", "ghost-door", "door-h", "door-v", "glasses"}
	for _, name := range names {
		if _, found := t.id[name]; !found {
			log.Fatalf("id %q does not exist", name)
		}
	}
}
