package main

import "github.com/go-gl/mathgl/mgl32"

type Ray struct {
	Orig, Dir mgl32.Vec3
	Rate      float32 // 运动物体的进度 0～1
}

func NewRay(orig mgl32.Vec3, dir mgl32.Vec3, rate float32) *Ray {
	return &Ray{Orig: orig, Dir: dir, Rate: rate}
}

func (r *Ray) At(rate float32) mgl32.Vec3 {
	return r.Orig.Add(r.Dir.Mul(rate))
}

func (r *Ray) Reflect(point mgl32.Vec3, normal mgl32.Vec3) *Ray {
	dir := r.Dir.Sub(normal.Mul(2 * r.Dir.Dot(normal))).Normalize()
	return NewRay(point, dir, r.Rate)
}

func (r *Ray) Refract(point mgl32.Vec3, normal mgl32.Vec3, refract float32) *Ray {
	cos := normal.Dot(r.Dir.Mul(-1))
	temp1 := r.Dir.Add(normal.Mul(cos)).Mul(refract)
	temp2 := normal.Mul(-Sqrt(mgl32.Abs(1 - temp1.LenSqr())))
	dir := temp1.Add(temp2).Normalize()
	return NewRay(point, dir, r.Rate)
}
