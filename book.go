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
	world       = NewHitList()
	background  = mgl32.Vec3{}
	sampleCount = float32(1)
)

func TestBook() {
	// 屏幕大小与视野角度
	w, h := float32(1280), float32(720)
	angle := mgl32.DegToRad(45)
	// 相机距离视口的距离
	l := w / 2 / Tan(angle)
	center := mgl32.Vec3{0, -100, 0} // 相机位置
	rotate := mgl32.HomogRotate3DX(mgl32.DegToRad(0)).
		Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(0))) // 相机旋转 x y 是反的
	// 世界设置
	background = mgl32.Vec3{0.5, 0.7, 1.0}
	// 设置地面
	world.Add(NewSphere(mgl32.Vec3{0, 10000, 600}, 10000, NewSolidLambert(mgl32.Vec3{0.5, 0.5, 0.5})))
	// 设置其他场景
	for i := -8; i < 8; i++ {
		for j := -5; j < 5; j++ {
			pos := mgl32.Vec3{float32(i)*60 + 20*rand.Float32(), -20, 600 + float32(j)*60 + 20*rand.Float32()}
			temp := rand.Float32()
			if temp < 0.8 {
				world.Add(NewSphere(pos, 20, NewSolidLambert(RandClr())))
			} else if temp < 0.95 {
				world.Add(NewSphere(pos, 20, NewMetal(RandVec(0.5, 1), Rand(0, 0.5))))
			} else {
				world.Add(NewSphere(pos, 20, NewDielectric(1.5)))
			}
		}
	}
	world.Add(NewSphere(mgl32.Vec3{0, -100, 600}, 100, NewDielectric(1.5)))
	world.Add(NewSphere(mgl32.Vec3{-200, -100, 600}, 100, NewSolidLambert(mgl32.Vec3{0.4, 0.2, 0.1})))
	world.Add(NewSphere(mgl32.Vec3{200, -100, 600}, 100, NewMetal(mgl32.Vec3{0.7, 0.6, 0.5}, 0)))
	// 进行渲染
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := 0; y < int(h); y++ {
		fmt.Println("new line", y)
		for x := 0; x < int(w); x++ {
			if x == 630 && y == 254 {
				fmt.Println("AAA")
			}
			clr := mgl32.Vec3{}
			if sampleCount > 1 { // 进行多次采样
				for i := 0; i < int(sampleCount); i++ {
					tx := float32(x) + rand.Float32() - 0.5
					ty := float32(y) + rand.Float32() - 0.5
					// 计算视口点的位置 坐标系并没有正规化
					vp := mgl32.Vec3{tx - w/2 + 0.5, ty - h/2 + 0.5, l}.Normalize()
					vp = rotate.Mat3().Mul3x1(vp) // 应用旋转，射线平移体现在相机位置上，射线方向不包含位置
					pos := rotate.Mul4x1(center.Vec4(1)).Vec3()
					ray := NewRay(pos, vp, rand.Float32())
					clr = clr.Add(RayColor(ray, 0))
				}
				clr = clr.Mul(1 / sampleCount)
			} else {
				// 计算视口点的位置 坐标系并没有正规化
				vp := mgl32.Vec3{float32(x) - w/2 + 0.5, float32(y) - h/2 + 0.5, l}.Normalize()
				vp = rotate.Mat3().Mul3x1(vp) // 应用旋转，射线平移体现在相机位置上，射线方向不包含位置
				pos := rotate.Mul4x1(center.Vec4(1)).Vec3()
				ray := NewRay(pos, vp, rand.Float32())
				clr = RayColor(ray, 0)
			}
			if NearZero(clr) {
				fmt.Println("clr is near zero")
			}
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
	hit := world.Hit(ray)
	if hit == nil { // 没有碰撞到返回背景
		return background
	}
	emitClr := hit.Material.Emitted(hit.UV, hit.Normal) // 发光色
	scatter := hit.Material.Scatter(ray, hit)
	if scatter == nil { // 不散射直接返回发光
		return emitClr
	}
	return Mul(scatter.Color, RayColor(scatter.Ray, dep+1)).Add(emitClr)
}
