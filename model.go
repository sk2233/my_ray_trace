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
	UV       mgl32.Vec2
}

type IHit interface {
	Hit(ray *Ray) *HitDetail
}

type Sphere struct {
	Center1, Center2 mgl32.Vec3
	Radius           float32
	Material         IMaterial
}

func NewSphere(center mgl32.Vec3, radius float32, material IMaterial) *Sphere {
	return &Sphere{Center1: center, Center2: center, Radius: radius, Material: material}
}

func NewMoveSphere(center1, center2 mgl32.Vec3, radius float32, material IMaterial) *Sphere {
	return &Sphere{Center1: center1, Center2: center2, Radius: radius, Material: material}
}

func (s *Sphere) Hit(ray *Ray) *HitDetail {
	center := s.GetCenter(ray.Rate)
	l := center.Sub(ray.Orig)
	tca := l.Dot(ray.Dir)
	if tca < 0 {
		return nil
	}
	d2 := l.LenSqr() - tca*tca
	if d2 > s.Radius*s.Radius {
		return nil
	}
	thc := Sqrt(s.Radius*s.Radius - d2)
	rate := tca - thc  // 优先选择最近的
	if rate <= 0.001 { // 射线的反向，尝试另一个交点
		rate = tca + thc
	}
	if rate <= 0.001 { // 仍然反向，没有交点
		return nil
	}
	point := ray.At(rate)
	normal := point.Sub(center).Normalize()
	uv := GetSphereUV(normal)
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
		UV:       uv,
	}
}

func (s *Sphere) GetCenter(rate float32) mgl32.Vec3 {
	return MixVec(s.Center1, s.Center2, rate)
}

type Quad struct {
	Q, U, V, W mgl32.Vec3
	Normal     mgl32.Vec3
	D          float32
	Material   IMaterial
	IsTriangle bool // 三角形可以当做特殊的 Quad 处理
}

func (q *Quad) Hit(ray *Ray) *HitDetail {
	d := q.Normal.Dot(ray.Dir)
	if d < MinOff { // 垂直于法线
		return nil
	}
	rate := (q.D - q.Normal.Dot(ray.Orig)) / d
	if rate <= MinOff { // 方向不对
		return nil
	}
	point := ray.At(rate)
	temp := point.Sub(q.Q)
	uv := mgl32.Vec2{ // 转换为四边形的 uv 值
		q.W.Dot(temp.Cross(q.V)),
		q.W.Dot(q.U.Cross(temp)),
	} // 判断是否在四边形范围内
	if uv[0] < 0 || uv[0] > 1 || uv[1] < 0 || uv[1] > 1 {
		return nil
	} // 三角形与四边形的主要区别
	if q.IsTriangle && uv[0]+uv[1] > 1 {
		return nil
	}
	outside := ray.Dir.Dot(q.Normal) < 0
	normal := q.Normal
	if !outside {
		normal = normal.Mul(-1)
	}
	return &HitDetail{
		Rate:     rate,
		Point:    point,
		Normal:   normal,
		Outside:  outside,
		Material: q.Material,
		UV:       uv,
	}
}

func NewQuad(q mgl32.Vec3, u mgl32.Vec3, v mgl32.Vec3, material IMaterial) *Quad {
	temp := u.Cross(v)
	normal := temp.Normalize()
	d := normal.Dot(q)
	w := temp.Mul(1 / temp.LenSqr())
	return &Quad{Q: q, U: u, V: v, W: w, Material: material, Normal: normal, D: d, IsTriangle: false}
}

func NewTriangle(q mgl32.Vec3, u mgl32.Vec3, v mgl32.Vec3, material IMaterial) *Quad {
	temp := u.Cross(v)
	normal := temp.Normalize()
	d := normal.Dot(q)
	w := temp.Mul(1 / temp.LenSqr())
	return &Quad{Q: q, U: u, V: v, W: w, Material: material, Normal: normal, D: d, IsTriangle: true}
}

type HitList struct {
	Hits []IHit
}

func NewHitList() *HitList {
	return &HitList{Hits: make([]IHit, 0)}
}

func (h *HitList) Hit(ray *Ray) *HitDetail {
	var res *HitDetail
	for _, hit := range h.Hits {
		temp := hit.Hit(ray)
		if temp == nil {
			continue
		}
		if res != nil && res.Rate < temp.Rate {
			continue
		}
		res = temp
	}
	return res
}

func (h *HitList) Add(hit IHit) {
	h.Hits = append(h.Hits, hit)
}
