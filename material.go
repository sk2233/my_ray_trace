package main

import "github.com/go-gl/mathgl/mgl32"

type ScatterDetail struct {
	Ray   *Ray
	Color mgl32.Vec3
}

type IMaterial interface {
	Scatter(ray *Ray, detail *HitDetail) *ScatterDetail
}

type Lambert struct {
	Texture ITexture
}

func NewSolidLambert(clr mgl32.Vec3) *Lambert {
	return &Lambert{Texture: NewSolidColor(clr)}
}

func NewLambert(texture ITexture) *Lambert {
	return &Lambert{Texture: texture}
}

func (l *Lambert) Scatter(ray *Ray, detail *HitDetail) *ScatterDetail {
	//dir := detail.Normal.Add(RandVec()) // 随机散射与法线加权
	//if dir.LenSqr() < MinOff {          // 太小了
	//	dir = detail.Normal
	//} else {
	//	dir = dir.Normalize()
	//}
	//ray = NewRay(detail.Point, dir, ray.Rate)
	//scale := max(dir.Dot(detail.Normal), 0)
	// 直接使用光栅图形了
	scale := ray.Dir.Mul(-1).Dot(detail.Normal)
	clr := l.Texture.Sample(detail.UV, detail.Point)
	return &ScatterDetail{
		Color: clr.Mul(scale),
	}
}

type Metal struct {
	Albedo mgl32.Vec3
	Fuzz   float32
}

func NewMetal(albedo mgl32.Vec3, fuzz float32) *Metal {
	return &Metal{Albedo: albedo, Fuzz: fuzz}
}

func (m *Metal) Scatter(ray *Ray, detail *HitDetail) *ScatterDetail {
	ray = ray.Reflect(detail.Point, detail.Normal)
	if m.Fuzz > 0 {
		ray.Dir = ray.Dir.Add(RandVec().Mul(m.Fuzz)).Normalize()
	}
	if ray.Dir.Dot(detail.Normal) <= 0 {
		return nil
	}
	return &ScatterDetail{
		Ray:   ray,
		Color: m.Albedo,
	}
}

type Dielectric struct {
	Refract float32 // 折射率
}

func (d *Dielectric) Scatter(ray *Ray, detail *HitDetail) *ScatterDetail {
	refract := d.Refract
	if detail.Outside {
		refract = 1 / d.Refract
	}
	cos := detail.Normal.Dot(ray.Dir.Mul(-1))
	sin := Sqrt(1 - cos*cos)
	if refract*sin > 1 || d.Reflectance(cos, refract) { // 只能反射
		ray = ray.Reflect(detail.Point, detail.Normal)
	} else { // 可以折射
		ray = ray.Refract(detail.Point, detail.Normal, refract)
	}
	return &ScatterDetail{
		Ray:   ray,
		Color: mgl32.Vec3{1, 1, 1},
	}
}

func (d *Dielectric) Reflectance(cos float32, refract float32) bool {
	r0 := (1 - refract) / (1 + refract) // 近似玻璃极端角度的镜面反射
	r0 = r0 * r0
	return r0+(1-r0)*Pow(1-cos, 5) > 0.5
}

func NewDielectric(refract float32) *Dielectric {
	return &Dielectric{Refract: refract}
}
