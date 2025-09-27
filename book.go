package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"sync"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
)

// https://raytracing.github.io/books/RayTracingInOneWeekend.html
// https://raytracing.github.io/books/RayTracingTheNextWeek.html

var (
	world       = NewHitList()
	background  = mgl32.Vec3{}
	sampleCount = float32(250) // 增加采样次数以减少噪点
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
	ground := NewSolidLambert(mgl32.Vec3{0.48, 0.83, 0.53})
	quads := make([]*Quad, 0)
	for i := -10; i < 10; i++ {
		for j := -4; j < 4; j++ {
			pos := mgl32.Vec3{float32(i) * 50, 150 + Rand(1, 50), 600 + float32(j)*50}
			temp := NewBox(mgl32.Vec3{50, 50, 50}, pos, mgl32.Vec3{}, ground)
			quads = append(quads, temp[0:4]...) // 下标 4 对应的是底部，可以不要
			quads = append(quads, temp[5])
		}
	}
	for _, quad := range quads {
		world.Add(quad)
	}
	//world.Add(NewMoveSphere(mgl32.Vec3{-100, -100, 600}, mgl32.Vec3{100, -100, 600}, 100, NewSolidLambert(mgl32.Vec3{0.7, 0.3, 0.1})))
	world.Add(NewSphere(mgl32.Vec3{-150, 100, 600}, 100, NewDielectric(1.5)))
	world.Add(NewSphere(mgl32.Vec3{150, 100, 600}, 100, NewMetal(mgl32.Vec3{0.8, 0.8, 0.9}, 0.1)))
	imgTex := NewLambert(NewImageTexture(LoadImage("res/test2.png")))
	world.Add(NewSphere(mgl32.Vec3{-400, -100, 600}, 150, imgTex))
	light := NewSolidLight(mgl32.Vec3{7, 7, 7})
	world.Add(NewQuad(mgl32.Vec3{-100, -300, 600 + 100}, mgl32.Vec3{0, 0, -200}, mgl32.Vec3{200, 0, 0}, light))
	// 进行渲染
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	total := int(w) * int(h)
	processed := int64(0)

	wg := sync.WaitGroup{}
	wg.Add(total)
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			go func(ax, ay int) {
				clr := mgl32.Vec3{}
				// 进行多次采样以减少噪点
				for i := 0; i < int(sampleCount); i++ {
					// 在像素内随机采样
					tx := float32(ax) + rand.Float32() - 0.5
					ty := float32(ay) + rand.Float32() - 0.5
					// 计算视口点的位置 坐标系并没有正规化
					vp := mgl32.Vec3{tx - w/2 + 0.5, ty - h/2 + 0.5, l}.Normalize()
					vp = rotate.Mat3().Mul3x1(vp) // 应用旋转，射线平移体现在相机位置上，射线方向不包含位置
					pos := rotate.Mul4x1(center.Vec4(1)).Vec3()
					ray := NewRay(pos, vp, rand.Float32())
					clr = clr.Add(RayColor(ray, 0))
				}
				clr = clr.Mul(1 / sampleCount) // 平均化采样结果
				clr = GammaAdjust(clr)         // gamma 矫正， 入参范围  0 ~ 1
				img.Set(ax, ay, color.RGBA{
					R: uint8(min(clr[0]*255, 255)),
					G: uint8(min(clr[1]*255, 255)),
					B: uint8(min(clr[2]*255, 255)),
					A: 255,
				})
				wg.Done()

				// 每处理1000个像素显示一次进度
				atomic.AddInt64(&processed, 1)
				temp := atomic.LoadInt64(&processed)
				if temp%10000 == 0 {
					progress := float64(temp) / float64(total) * 100
					fmt.Printf("渲染进度: %.1f%% (%d/%d)\n", progress, temp, total)
				}
			}(x, y)
		}
	}
	wg.Wait()
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
