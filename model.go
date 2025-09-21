package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

type HitDetail struct {
	Rate     float32
	Point    mgl32.Vec3
	Normal   mgl32.Vec3
	Outside  bool
	Material IMaterial
}

type IHit interface {
	Hit(ray *Ray) *HitDetail
}

type Sphere struct {
	Center   mgl32.Vec3
	Radius   float32
	Material IMaterial
}

func NewSphere(center mgl32.Vec3, radius float32) *Sphere {
	return &Sphere{Center: center, Radius: radius}
}

func (s *Sphere) Hit(ray *Ray) *HitDetail {
	l := s.Center.Sub(ray.Orig)
	tca := l.Dot(ray.Dir)
	if tca < 0 {
		return nil
	}
	d2 := l.LenSqr() - tca*tca
	if d2 > s.Radius*s.Radius {
		return nil
	}
	thc := Sqrt(s.Radius*s.Radius - d2)
	rate := tca - thc   // 优先选择最近的
	if rate <= MinOff { // 射线的反向，尝试另一个交点
		rate = tca + thc
	}
	if rate <= MinOff { // 仍然反向，没有交点
		return nil
	}
	point := ray.At(rate)
	normal := point.Sub(s.Center).Normalize()
	outside := ray.Dir.Dot(normal) < 0
	if !outside { // 在外面的话注意翻转法线
		normal = normal.Mul(-1)
	}
	return &HitDetail{
		Rate:     rate,
		Point:    point,
		Normal:   normal,
		Outside:  outside,
		Material: s.Material,
	}
}

type HitList struct {
	Hits []IHit
}

func NewHitList() *HitList {
	return &HitList{Hits: make([]IHit, 0)}
}

func (h *HitList) Hit(ray *Ray) *HitDetail {
	for _, hit := range h.Hits {
		if res := hit.Hit(ray); res != nil {
			return res
		}
	}
	return nil
}

func (h *HitList) Add(hit IHit) {
	h.Hits = append(h.Hits, hit)
}

type Ray struct {
	Orig, Dir mgl32.Vec3
}

func NewRay(orig mgl32.Vec3, dir mgl32.Vec3) *Ray {
	return &Ray{Orig: orig, Dir: dir}
}

func (r *Ray) At(rate float32) mgl32.Vec3 {
	return r.Orig.Add(r.Dir.Mul(rate))
}

func (r *Ray) Reflect(point mgl32.Vec3, normal mgl32.Vec3) *Ray {
	dir := r.Dir.Sub(normal.Mul(2 * r.Dir.Dot(normal))).Normalize()
	return NewRay(point, dir)
}

type ScatterDetail struct {
	Ray   *Ray
	Color mgl32.Vec3
}

type IMaterial interface {
	Scatter(ray *Ray, detail *HitDetail) *ScatterDetail
}

type Lambert struct {
	Albedo mgl32.Vec3
}

func (l *Lambert) Scatter(ray *Ray, detail *HitDetail) *ScatterDetail {
	dir := detail.Normal.Add(RandVec()).Normalize() // 随机散射与法线加权
	ray = NewRay(detail.Point, dir)
	return &ScatterDetail{
		Ray:   ray,
		Color: l.Albedo,
	}
}

func NewLambert(albedo mgl32.Vec3) *Lambert {
	return &Lambert{Albedo: albedo}
}

type Metal struct {
	Albedo mgl32.Vec3
}

func NewMetal(albedo mgl32.Vec3) *Metal {
	return &Metal{Albedo: albedo}
}

func (m *Metal) Scatter(ray *Ray, detail *HitDetail) *ScatterDetail {
	ray = ray.Reflect(detail.Point, detail.Normal)
	return &ScatterDetail{
		Ray:   ray,
		Color: m.Albedo,
	}
}
