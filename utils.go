package main

import (
	"image"
	"image/png"
	"math"
	"math/rand/v2"
	"os"

	"github.com/go-gl/mathgl/mgl32"
)

// 注意  v1 v2 的顺序
func Mix(v1, v2, rate float32) float32 {
	return v1*rate + v2*(1-rate)
}

func MixVec(v1, v2 mgl32.Vec3, rate float32) mgl32.Vec3 {
	return mgl32.Vec3{Mix(v1[0], v2[0], rate), Mix(v1[1], v2[1], rate), Mix(v1[2], v2[2], rate)}
}

func Sqrt(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func Tan(v float32) float32 {
	return float32(math.Tan(float64(v)))
}

func Cos(v float32) float32 {
	return float32(math.Cos(float64(v)))
}

func Pow(x float32, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}

func Mul(v1, v2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{v1[0] * v2[0], v1[1] * v2[1], v1[2] * v2[2]}
}

func Max(v1, v2 float32) float32 {
	if v1 > v2 {
		return v1
	}
	return v2
}

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func SaveImage(image image.Image, filename string) {
	file, err := os.Create(filename)
	HandleErr(err)
	defer file.Close()
	err = png.Encode(file, image)
	HandleErr(err)
}

func GammaAdjust(v mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		Sqrt(v[0]),
		Sqrt(v[1]),
		Sqrt(v[2]),
	}
}

func RandClr() mgl32.Vec3 {
	return mgl32.Vec3{
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
	}
}

func RandVec() mgl32.Vec3 {
	res := mgl32.Vec3{
		float32(rand.NormFloat64()),
		float32(rand.NormFloat64()),
		float32(rand.NormFloat64()),
	}
	return res.Normalize()
}

func Acos(val float32) float32 {
	return float32(math.Acos(float64(val)))
}

func Atan2(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

func GetSphereUV(normal mgl32.Vec3) mgl32.Vec2 {
	th := Acos(-normal[1])
	ph := Atan2(-normal[2], normal[0]) + math.Pi
	return mgl32.Vec2{
		ph / (2 * math.Pi),
		th / math.Pi,
	}
}

func LoadImage(path string) image.Image {
	file, err := os.Open(path)
	HandleErr(err)
	defer file.Close()
	img, err := png.Decode(file)
	HandleErr(err)
	return img
}

// vs 盒子的8个顶点
func NewBox(size, pos, rotate mgl32.Vec3, material IMaterial) []*Quad {
	scal := mgl32.Scale3D(size[0], size[1], size[2])
	tran := mgl32.Translate3D(pos[0], pos[1], pos[2])
	rota := mgl32.HomogRotate3DX(rotate[0]).Mul4(mgl32.HomogRotate3DY(rotate[1])).Mul4(mgl32.HomogRotate3DZ(rotate[2]))
	mat := tran.Mul4(rota).Mul4(scal)
	vs := []mgl32.Vec3{{0, 0, 1}, {1, 0, 1}, {1, 0, 0}, {0, 0, 0}, {0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}}
	for i := 0; i < len(vs); i++ {
		vs[i] = mat.Mul4x1(vs[i].Vec4(1)).Vec3()
	}
	res := make([]*Quad, 0)
	res = append(res, NewQuad(vs[3], vs[0].Sub(vs[3]), vs[2].Sub(vs[3]), material))
	res = append(res, NewQuad(vs[3], vs[2].Sub(vs[3]), vs[7].Sub(vs[3]), material))
	res = append(res, NewQuad(vs[3], vs[7].Sub(vs[3]), vs[0].Sub(vs[3]), material))
	res = append(res, NewQuad(vs[5], vs[6].Sub(vs[5]), vs[1].Sub(vs[5]), material))
	res = append(res, NewQuad(vs[5], vs[4].Sub(vs[5]), vs[6].Sub(vs[5]), material))
	res = append(res, NewQuad(vs[5], vs[1].Sub(vs[5]), vs[4].Sub(vs[5]), material))
	return res
}
