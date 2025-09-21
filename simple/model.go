package simple

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Sphere struct {
	Center                      mgl32.Vec3
	Radius                      float32
	SurfaceColor, EmissionColor mgl32.Vec3
	Transparency                float32
	Reflection                  bool
}

func (s *Sphere) Intersect(orig mgl32.Vec3, dir mgl32.Vec3) (bool, float32, float32) {
	l := s.Center.Sub(orig) // 斜边
	tca := l.Dot(dir)       // 在射线方向的投影，一条直角边
	if tca < 0 {            // 反反向不相交
		return false, 0, 0
	}
	d2 := l.LenSqr() - tca*tca  // 另一条直角边
	if d2 > s.Radius*s.Radius { // 距圆心太远
		return false, 0, 0
	}
	thc := Sqrt(s.Radius*s.Radius - d2) // 球内线段半距离
	return true, tca - thc, tca + thc
}
