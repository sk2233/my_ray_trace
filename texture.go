package main

import (
	"image"

	"github.com/go-gl/mathgl/mgl32"
)

type ITexture interface {
	Sample(uv mgl32.Vec2, pos mgl32.Vec3) mgl32.Vec3
}

type SolidColor struct {
	Albedo mgl32.Vec3
}

func (s *SolidColor) Sample(_ mgl32.Vec2, _ mgl32.Vec3) mgl32.Vec3 {
	return s.Albedo
}

func NewSolidColor(color mgl32.Vec3) *SolidColor {
	return &SolidColor{Albedo: color}
}

type ChessTexture struct {
	Size      float32
	Even, Odd ITexture
}

func (c *ChessTexture) Sample(uv mgl32.Vec2, pos mgl32.Vec3) mgl32.Vec3 {
	x := int((pos[0] + 1000) / c.Size) // 避免正负数交点截断问题
	y := int((pos[1] + 1000) / c.Size)
	z := int((pos[2] + 1000) / c.Size)
	if (x+y+z)%2 == 0 {
		return c.Even.Sample(uv, pos)
	} else {
		return c.Odd.Sample(uv, pos)
	}
}

func NewChessTexture(size float32, even ITexture, odd ITexture) *ChessTexture {
	return &ChessTexture{Size: size, Even: even, Odd: odd}
}

type ImageTexture struct {
	Image         *image.RGBA
	Width, Height float32
}

func (i *ImageTexture) Sample(uv mgl32.Vec2, pos mgl32.Vec3) mgl32.Vec3 {
	// 默认采用循环采样的方式
	uv[0] = uv[0] - float32(int(uv[0]))
	uv[1] = uv[1] - float32(int(uv[1]))
	x := int(i.Width * uv[0])
	y := int(i.Height * uv[1])
	clr := i.Image.RGBAAt(x, y)
	return mgl32.Vec3{
		float32(clr.R) / 255,
		float32(clr.G) / 255,
		float32(clr.B) / 255,
	}
}

func NewImageTexture(image *image.RGBA) *ImageTexture {
	bound := image.Bounds()
	return &ImageTexture{Image: image, Width: float32(bound.Dx()), Height: float32(bound.Dy())}
}
