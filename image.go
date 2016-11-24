package main

import (
	"image"
	idraw "image/draw"
	"log"
	"os"
	"path/filepath"

	"github.com/qeedquan/go-media/image/imageutil"
	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

type Image struct {
	tex  *sdl.Texture
	size sdl.Point
}

func loadImage(name string, filters ...func(image.Image) image.Image) *Image {
	log.SetPrefix("image: ")
	filename := filepath.Join(assets, name)
	f, err := os.Open(filename)
	ck(err)
	defer f.Close()

	img, _, err := image.Decode(f)
	ck(err)

	for _, filter := range filters {
		img = filter(img)
	}

	return makeImage(img)
}

func colorKey(key sdl.Color) func(image.Image) image.Image {
	return func(img image.Image) image.Image {
		return imageutil.ColorKey(img, key)
	}
}

func makeImage(img image.Image) *Image {
	log.SetPrefix("image: ")

	r := img.Bounds()
	size := sdl.Point{int32(r.Dx()), int32(r.Dy())}

	tex, err := screen.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_TARGET, int(size.X), int(size.Y))
	ck(err)

	rgba := image.NewRGBA(image.Rect(0, 0, int(size.X), int(size.Y)))
	idraw.Draw(rgba, rgba.Bounds(), img, image.ZP, idraw.Src)

	tex.Update(nil, rgba.Pix[:], rgba.Stride)
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)

	return &Image{
		tex:  tex,
		size: size,
	}
}

func makeTexture(size sdl.Point) *Image {
	log.SetPrefix("image: ")
	tex, err := screen.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_TARGET, int(size.X), int(size.Y))
	ck(err)

	tex.SetBlendMode(sdl.BLENDMODE_BLEND)

	return &Image{
		tex:  tex,
		size: size,
	}
}

func loadRGBA(name string) *image.RGBA {
	log.SetPrefix("image: ")
	filename := filepath.Join(assets, name)
	f, err := os.Open(filename)
	ck(err)
	defer f.Close()

	img, _, err := image.Decode(f)
	ck(err)

	r := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	idraw.Draw(rgba, rgba.Bounds(), img, image.ZP, idraw.Src)
	return rgba
}

func (m *Image) Draw(pos sdl.Point) {
	screen.Copy(m.tex, nil, &sdl.Rect{pos.X, pos.Y, m.size.X, m.size.Y})
}

func (m *Image) Bind() {
	log.SetPrefix("image: ")
	err := screen.SetTarget(m.tex)
	if err != nil {
		panic(err)
	}
}

func (m *Image) Unbind() {
	log.SetPrefix("image: ")
	err := screen.SetTarget(nil)
	if err != nil {
		panic(err)
	}
}

func (m *Image) Free() {
	m.tex.Destroy()
}

func loadFont(name string, ptsize int) *sdlttf.Font {
	log.SetPrefix("font: ")
	filename := filepath.Join(assets, name)
	font, err := sdlttf.OpenFont(filename, ptsize)
	ck(err)
	return font
}
