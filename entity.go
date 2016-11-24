package main

import "github.com/qeedquan/go-media/sdl"

type Entity struct {
	pos     sdl.Point
	vel     sdl.Point
	home    sdl.Point
	nearest sdl.Point
	speed   int
	state   int
}
