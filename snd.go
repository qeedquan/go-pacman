package main

import (
	"log"
	"path/filepath"

	"github.com/qeedquan/go-media/sdl/sdlmixer"
)

type Snd struct {
	pellet      [2]*sdlmixer.Chunk
	powerPellet *sdlmixer.Chunk
	fruitBounce *sdlmixer.Chunk
	eatgh       *sdlmixer.Chunk
	eatFruit    *sdlmixer.Chunk
	extraLife   *sdlmixer.Chunk
}

func newSnd() *Snd {
	return &Snd{
		pellet: [2]*sdlmixer.Chunk{
			loadSnd("pellet1.wav"),
			loadSnd("pellet2.wav"),
		},
		powerPellet: loadSnd("powerpellet.wav"),
		eatgh:       loadSnd("eatgh2.wav"),
		fruitBounce: loadSnd("fruitbounce.wav"),
		eatFruit:    loadSnd("eatfruit.wav"),
		extraLife:   loadSnd("extralife.wav"),
	}
}

func loadSnd(name string) *sdlmixer.Chunk {
	log.SetPrefix("sound: ")
	filename := filepath.Join(assets, "sounds", name)
	chunk, err := sdlmixer.LoadWAV(filename)
	if err != nil {
		log.Print(err)
		return nil
	}
	return chunk
}

func playSnd(chunk *sdlmixer.Chunk) {
	if !sfx || chunk == nil {
		return
	}
	chunk.PlayChannel(-1, 0)
}
