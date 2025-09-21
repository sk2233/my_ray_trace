package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
)

// https://raytracing.github.io/books/RayTracingInOneWeekend.html

var (
	world = NewHitList()
)

func TestBook() {
	// 屏幕大小与视野角度
	w, h := float32(1280), float32(720)
	angle := mgl32.DegToRad(30)
	// 相机距离视口的距离
	l := w / 2 / Tan(angle)
	center := mgl32.Vec3{} // 相机位置
	// 世界设置
	world.Add(NewSphere(mgl32.Vec3{0, 0, 1000}, 200))
	world.Add(NewSphere(mgl32.Vec3{0, 2200, 1000}, 2000))
	// 进行渲染
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := 0; y < int(h); y++ {
		fmt.Println("new line", y)
		for x := 0; x < int(w); x++ {
			// 计算视口点的位置，假设相机位于原点  坐标系并没有正规化
			vp := mgl32.Vec3{float32(x) - w/2 + 0.5, float32(y) - h/2 + 0.5, l}
			ray := NewRay(center, vp.Sub(center).Normalize())
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
			return Mul(scatter.Color, RayColor(scatter.Ray, dep+1))
		}
		return mgl32.Vec3{} // 完全吸收光线
	}
	rate := (ray.Dir.Y() + 1) * 0.5
	return MixVec(mgl32.Vec3{1, 1, 1}, mgl32.Vec3{0.5, 0.7, 1}, rate)
}
