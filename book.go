package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"

	"github.com/go-gl/mathgl/mgl32"
)

// https://raytracing.github.io/books/RayTracingInOneWeekend.html
// https://raytracing.github.io/books/RayTracingTheNextWeek.html

var (
	world = NewHitList()
)

func TestBook() {
	// 屏幕大小与视野角度
	w, h := float32(1280), float32(720)
	angle := mgl32.DegToRad(45)
	// 相机距离视口的距离
	l := w / 2 / Tan(angle)
	center := mgl32.Vec3{0, 0, -100} // 相机位置
	rotate := mgl32.HomogRotate3DX(mgl32.DegToRad(0)).
		Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(0))) // 相机旋转 x y 是反的
	//fieldDep := float32(10)
	// 世界设置
	tempTexture := NewImageTexture(LoadImage("res/test.png"))
	for i := 0; i < 15; i++ {
		x := rand.Float32()*1280 - 640
		z := rand.Float32()*640 - 320 + 600
		temp := rand.Float32()
		if temp < 0.3 {
			world.Add(NewSphere(mgl32.Vec3{x, 150, z}, 50, NewDielectric(1.5)))
		} else if temp < 0.6 {
			world.Add(NewSphere(mgl32.Vec3{x, 150, z}, 50, NewMetal(RandClr(), 0)))
		} else {
			world.Add(NewSphere(mgl32.Vec3{x, 150, z}, 50, NewLambert(tempTexture)))
		}
	}
	world.Add(NewSphere(mgl32.Vec3{-200, 0, 600}, 200, NewDielectric(1.5)))
	world.Add(NewSphere(mgl32.Vec3{200, 0, 600}, 200, NewMetal(mgl32.Vec3{0.8, 0.6, 0.2}, 0)))
	//world.Add(NewMoveSphere(mgl32.Vec3{-50, -400, 600}, mgl32.Vec3{50, -400, 600}, 200, NewMetal(mgl32.Vec3{1.0, 0.2, 0.2}, 0)))
	texture := NewChessTexture(50, NewSolidColor(mgl32.Vec3{0.2, 0.3, 0.1}), NewSolidColor(mgl32.Vec3{0.9, 0.9, 0.9}))
	//world.Add(NewSphere(mgl32.Vec3{0, 10200, 600}, 10000, NewLambert(texture)))
	world.Add(NewQuad(mgl32.Vec3{-640, 200, 600 - 360}, mgl32.Vec3{0, 0, 720}, mgl32.Vec3{1280, 0, 0}, NewLambert(texture)))
	// 进行渲染
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := 0; y < int(h); y++ {
		fmt.Println("new line", y)
		for x := 0; x < int(w); x++ {
			// 计算视口点的位置 坐标系并没有正规化
			vp := mgl32.Vec3{float32(x) - w/2 + 0.5, float32(y) - h/2 + 0.5, l}.Normalize()
			//fieldWeight := 1 - Pow(vp.Dot(mgl32.Vec3{0, 0, 1}), 5) // 越靠近中心越清晰
			vp = rotate.Mat3().Mul3x1(vp) // 应用旋转，射线平移体现在相机位置上，射线方向不包含位置
			//pos := center.Add(mgl32.Vec3{rand.Float32(), rand.Float32(), 0}.Mul(fieldDep * fieldWeight))
			pos := center
			pos = rotate.Mul4x1(pos.Vec4(1)).Vec3()
			ray := NewRay(pos, vp, rand.Float32())
			clr := RayColor(ray, 0)
			clr = GammaAdjust(clr) // gamma 矫正， 入参范围  0 ~ 1
			img.Set(x, y, color.RGBA{
				R: uint8(min(clr[0]*255, 255)),
				G: uint8(min(clr[1]*255, 255)),
				B: uint8(min(clr[2]*255, 255)),
				A: 255,
			})
		}
	}
	fmt.Println("all over")
	SaveImage(img, "output/book.png")
}

func RayColor(ray *Ray, dep int) mgl32.Vec3 {
	if dep >= MaxDep { // 防止无限递归
		return mgl32.Vec3{}
	}
	if hit := world.Hit(ray); hit != nil { // 碰撞到物体
		if scatter := hit.Material.Scatter(ray, hit); scatter != nil { // 进行散射
			if scatter.Ray == nil { // 直接给颜色没有散射
				return scatter.Color
			}
			return Mul(scatter.Color, RayColor(scatter.Ray, dep+1))
		}
		return mgl32.Vec3{} // 完全吸收光线
	}
	rate := ray.Dir.Y() + 0.5
	return MixVec(mgl32.Vec3{1, 1, 1}, mgl32.Vec3{0.5, 0.7, 1}, rate)
}
