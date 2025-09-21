package simple

import (
	"image"
	"image/color"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// https://www.scratchapixel.com/lessons/3d-basic-rendering/introduction-to-ray-tracing/how-does-it-work.html

func TestSimple() {
	spheres := make([]*Sphere, 0)
	// 地面
	spheres = append(spheres, &Sphere{mgl32.Vec3{0.0, -10004, -20}, 10000, mgl32.Vec3{0.20, 0.20, 0.20},
		mgl32.Vec3{}, 0, false})
	// 物体
	spheres = append(spheres, &Sphere{mgl32.Vec3{0.0, 0, -20}, 4, mgl32.Vec3{1.00, 0.32, 0.36},
		mgl32.Vec3{}, 0.5, true})
	spheres = append(spheres, &Sphere{mgl32.Vec3{5.0, -1, -15}, 2, mgl32.Vec3{0.90, 0.76, 0.46},
		mgl32.Vec3{}, 0, true})
	spheres = append(spheres, &Sphere{mgl32.Vec3{5.0, 0, -25}, 3, mgl32.Vec3{0.65, 0.77, 0.97},
		mgl32.Vec3{}, 0, true})
	spheres = append(spheres, &Sphere{mgl32.Vec3{-5.5, 0, -15}, 3, mgl32.Vec3{0.90, 0.90, 0.90},
		mgl32.Vec3{}, 0, true})
	// 灯
	spheres = append(spheres, &Sphere{mgl32.Vec3{0.0, 20, -30}, 3, mgl32.Vec3{},
		mgl32.Vec3{3, 3, 3}, 0, false})
	img := Render(spheres, 1280, 720)
	SaveImage(img, "output/spheres.png")
}

func Render(spheres []*Sphere, w float32, h float32) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	aspect := w / h
	angle := Tan(30.0 / 180.0 * math.Pi / 2)
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			// 计算射线方向
			tx := ((float32(x)+0.5)/w*2 - 1) * angle * aspect
			ty := (1 - (float32(y)+0.5)/h*2) * angle
			rayDir := mgl32.Vec3{tx, ty, -1}.Normalize()
			res := Trace(mgl32.Vec3{}, rayDir, spheres, 0)
			img.Set(x, y, color.RGBA{
				R: uint8(min(255*res[0], 255)),
				G: uint8(min(255*res[1], 255)),
				B: uint8(min(255*res[2], 255)), A: 255})
		}
	}
	return img
}

func Trace(orig mgl32.Vec3, dir mgl32.Vec3, spheres []*Sphere, depth int) mgl32.Vec3 {
	// 先找最近相交的球体
	tnear := float32(math.MaxFloat32)
	var sphere *Sphere
	for i := 0; i < len(spheres); i++ {
		if hit, l0, l1 := spheres[i].Intersect(orig, dir); hit {
			if l0 < 0 { // l0 < l1
				l0 = l1
			}
			if l0 < tnear {
				tnear = l0
				sphere = spheres[i]
			}
		}
	}
	if sphere == nil { // 没有命中返回背景色
		return mgl32.Vec3{2, 2, 2}
	}
	phit := orig.Add(dir.Mul(tnear))            // 命中点
	nhit := phit.Sub(sphere.Center).Normalize() // 命中点的法线
	// 对命中点加基于法线的偏移确定其在内外的位置，在球面上因为精度问题可能随机算作球内或球外
	bias := float32(1e-4)
	inside := false
	if dir.Dot(nhit) > 0 { // 判断 orig 是否在球的内部
		nhit = nhit.Mul(-1) // 法线也要修改
		inside = true
	}
	// 判断是否还需要进行光线追踪
	if (sphere.Transparency > 0 || sphere.Reflection) && depth < MaxRayDepth {
		// 计算折射率
		fre := Mix(1, Pow(1+dir.Dot(nhit), 3), 0.1)
		// 计算反射光
		reflDir := dir.Sub(nhit.Mul(2 * dir.Dot(nhit))).Normalize() // 计算反射方向
		refl := Trace(phit.Add(nhit.Mul(bias)), reflDir, spheres, depth+1)
		// 计算折射光
		refr := mgl32.Vec3{}
		if sphere.Transparency > 0 { // 折射
			eta := float32(1.0 / 1.1)
			if inside {
				eta = 1.1
			}
			cosi := -nhit.Dot(dir)
			k := 1 - eta*eta*(1-cosi*cosi) // 计算折射方向
			refrDir := dir.Mul(eta).Add(nhit.Mul(eta*cosi - Sqrt(k))).Normalize()
			refr = Trace(phit.Add(nhit.Mul(bias)), refrDir, spheres, depth+1)
		}
		surfaceColor := Mul(refl.Mul(fre).Add(refr.Mul((1-fre)*sphere.Transparency)), sphere.SurfaceColor)
		return surfaceColor.Add(sphere.EmissionColor)
	}
	// 只受发光物体的影响
	surfaceColor := mgl32.Vec3{}
	for i := 0; i < len(spheres); i++ {
		emisColor := spheres[i].EmissionColor
		if emisColor[0] == 0 && emisColor[1] == 0 && emisColor[2] == 0 {
			continue // 非灯光对象
		}
		block := false
		lightDir := spheres[i].Center.Sub(phit).Normalize()
		for j := 0; j < len(spheres); j++ {
			if i == j {
				continue
			}
			if ok, _, _ := spheres[j].Intersect(phit.Add(nhit.Mul(bias)), lightDir); ok {
				block = true
				break
			}
		}
		if block { // 被其他物体遮挡了
			continue
		}
		surfaceColor = surfaceColor.Add(Mul(sphere.SurfaceColor, emisColor).Mul(Max(0, nhit.Dot(lightDir))))
	}
	return surfaceColor.Add(sphere.EmissionColor)
}
