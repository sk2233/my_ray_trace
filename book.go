package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/go-gl/mathgl/mgl32"
)

// https://raytracing.github.io/books/RayTracingInOneWeekend.html
// https://raytracing.github.io/books/RayTracingTheNextWeek.html

var (
	world       = NewHitList()
	background  = mgl32.Vec3{}
	sampleCount = float32(100) // 增加采样次数以减少噪点
)

func TestBook() {
	// 屏幕大小与视野角度
	w, h := float32(1280), float32(720)
	angle := mgl32.DegToRad(45)
	// 相机距离视口的距离
	l := w / 2 / Tan(angle)
	center := mgl32.Vec3{0, 0, 0} // 相机位置
	rotate := mgl32.HomogRotate3DX(mgl32.DegToRad(0)).
		Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(0))) // 相机旋转 x y 是反的
	// 世界设置
	background = mgl32.Vec3{0, 0, 0}
	red := NewSolidLambert(mgl32.Vec3{0.65, 0.05, 0.05})
	white := NewSolidLambert(mgl32.Vec3{0.73, 0.73, 0.73})
	green := NewSolidLambert(mgl32.Vec3{0.12, 0.45, 0.15})
	light := NewSolidLight(mgl32.Vec3{15, 15, 15})
	world.Add(NewQuad(mgl32.Vec3{-300, -300, 700 + 200}, mgl32.Vec3{0, 0, -400}, mgl32.Vec3{600, 0, 0}, white))
	world.Add(NewQuad(mgl32.Vec3{-300, -300, 700 + 200}, mgl32.Vec3{600, 0, 0}, mgl32.Vec3{0, 600, 0}, white))
	world.Add(NewQuad(mgl32.Vec3{-300, -300, 700 + 200}, mgl32.Vec3{0, 600, 0}, mgl32.Vec3{0, 0, -400}, green))
	world.Add(NewQuad(mgl32.Vec3{300, 300, 700 - 200}, mgl32.Vec3{-600, 0, 0}, mgl32.Vec3{0, 0, 400}, white))
	world.Add(NewQuad(mgl32.Vec3{300, 300, 700 - 200}, mgl32.Vec3{0, 0, 400}, mgl32.Vec3{0, -600, 0}, red))
	world.Add(NewQuad(mgl32.Vec3{-100, -300, 700 + 100}, mgl32.Vec3{0, 0, -200}, mgl32.Vec3{200, 0, 0}, light))
	// 两个旋转的方块
	quads := NewBox(mgl32.Vec3{100, 400, 100}, mgl32.Vec3{-200, -100, 700 - 100}, mgl32.Vec3{0, math.Pi / 12, 0}, white)
	quads = append(quads, NewBox(mgl32.Vec3{100, 200, 100}, mgl32.Vec3{100, 100, 700 - 100}, mgl32.Vec3{0, -math.Pi / 12, 0}, white)...)
	for _, quad := range quads {
		world.Add(quad)
	}
	// 进行渲染
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	total := int(w) * int(h)
	processed := 0

	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			clr := mgl32.Vec3{}
			// 进行多次采样以减少噪点
			for i := 0; i < int(sampleCount); i++ {
				// 在像素内随机采样
				tx := float32(x) + rand.Float32() - 0.5
				ty := float32(y) + rand.Float32() - 0.5
				// 计算视口点的位置 坐标系并没有正规化
				vp := mgl32.Vec3{tx - w/2 + 0.5, ty - h/2 + 0.5, l}.Normalize()
				vp = rotate.Mat3().Mul3x1(vp) // 应用旋转，射线平移体现在相机位置上，射线方向不包含位置
				pos := rotate.Mul4x1(center.Vec4(1)).Vec3()
				ray := NewRay(pos, vp, rand.Float32())
				clr = clr.Add(RayColor(ray, 0))
			}
			clr = clr.Mul(1 / sampleCount) // 平均化采样结果
			clr = GammaAdjust(clr)         // gamma 矫正， 入参范围  0 ~ 1
			img.Set(x, y, color.RGBA{
				R: uint8(min(clr[0]*255, 255)),
				G: uint8(min(clr[1]*255, 255)),
				B: uint8(min(clr[2]*255, 255)),
				A: 255,
			})

			// 每处理1000个像素显示一次进度
			processed++
			if processed%1000 == 0 {
				progress := float64(processed) / float64(total) * 100
				fmt.Printf("渲染进度: %.1f%% (%d/%d)\n", progress, processed, total)
			}
		}
	}
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
