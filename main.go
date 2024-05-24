package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"
)

var (
	outputFile = "sphere.gif"

	colorBlack = color.RGBA{A: 255}
	colorWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colorRed   = color.RGBA{R: 255, A: 255}
	colorGreen = color.RGBA{G: 255, A: 255}
	colorBlue  = color.RGBA{B: 255, A: 255}

	width  float64 = 400 // pixels
	height float64 = 400 // pixels
)

// Vertex merepresentasikan titik dalam ruang 3D.
type Vertex struct {
	x, y, z float64
}

// Triangle merepresentasikan segitiga dalam ruang 3D.
type Triangle struct {
	v1, v2, v3 Vertex
	color      color.Color
}

func main() {
	// Membuat bola dunia dengan jumlah garis lintang sebanyak 50 dan jumlah garis bujur sebanyak 50.
	tris := generateSphere(50, 50)

	// Menginisialisasi animasi GIF. Dengan loopCount = 0, maka animasi akan berulang terus menerus.
	anim := gif.GIF{LoopCount: 0}

	// Kita mulai dengan melakukan iterasi sebanyak 75 kali. Setiap iterasi akan membuat satu frame animasi.
	for i := 0; i < 75; i++ {
		// Untuk setiap frame, kita membuat gambar baru dengan lebar dan tinggi sesuai dengan nilai width dan height yang telah ditentukan sebelumnya.
		img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

		// Setiap pixel pada gambar akan diisi dengan warna hitam untuk background.
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(width); x++ {
				img.Set(x, y, color.Black)
			}
		}

		// Kita akan melakukan rotasi terhadap bola dunia dengan mengubah nilai heading dan pitch.
		heading := float64(i) * 2 * math.Pi / 75
		pitch := float64(i) * 2 * math.Pi / 75

		// Membuat matriks transformasi untuk rotasi.
		headingTransform := Matrix3{
			{math.Cos(heading), 0, -math.Sin(heading)},
			{0, 1, 0},
			{math.Sin(heading), 0, math.Cos(heading)},
		}
		pitchTransform := Matrix3{
			{1, 0, 0},
			{0, math.Cos(pitch), math.Sin(pitch)},
			{0, -math.Sin(pitch), math.Cos(pitch)},
		}

		// Menggabungkan matriks transformasi heading dan pitch.
		transform := headingTransform.multiply(pitchTransform)

		// Membuat z-buffer untuk menentukan pixel mana yang akan ditampilkan.
		zBuffer := make([]float64, int(width*height))
		for q := range zBuffer {
			zBuffer[q] = math.Inf(-1)
		}

		// Untuk setiap segitiga pada bola dunia, kita akan melakukan transformasi terhadap titik-titiknya.
		for _, t := range tris {

			// Melakukan transformasi terhadap titik-titik segitiga. Setelah itu, titik-titik tersebut akan dipindahkan ke tengah gambar.
			v1 := transform.transform(t.v1)
			v1.x += width / 2
			v1.y += height / 2
			v2 := transform.transform(t.v2)
			v2.x += width / 2
			v2.y += height / 2
			v3 := transform.transform(t.v3)
			v3.x += width / 2
			v3.y += height / 2

			// Menghitung normal dari segitiga tersebut.
			ab := Vertex{v2.x - v1.x, v2.y - v1.y, v2.z - v1.z}
			ac := Vertex{v3.x - v1.x, v3.y - v1.y, v3.z - v1.z}
			norm := Vertex{
				ab.y*ac.z - ab.z*ac.y,
				ab.z*ac.x - ab.x*ac.z,
				ab.x*ac.y - ab.y*ac.x,
			}

			// Normalisasi normal. Normalisasi adalah proses untuk mengubah panjang vektor menjadi 1.
			normalLength := math.Sqrt(norm.x*norm.x + norm.y*norm.y + norm.z*norm.z)
			norm.x /= normalLength
			norm.y /= normalLength
			norm.z /= normalLength

			// Menghitung sudut antara normal dan vektor pandang. Sudut ini akan digunakan untuk menentukan seberapa terang warna segitiga tersebut.
			angleCos := math.Abs(norm.z)

			// Menghitung area segitiga. Area segitiga ini akan digunakan untuk menentukan apakah suatu pixel berada di dalam segitiga atau tidak.
			minX := int(math.Max(0, math.Ceil(math.Min(v1.x, math.Min(v2.x, v3.x)))))
			maxX := int(math.Min(width-1, math.Floor(math.Max(v1.x, math.Max(v2.x, v3.x)))))
			minY := int(math.Max(0, math.Ceil(math.Min(v1.y, math.Min(v2.y, v3.y)))))
			maxY := int(math.Min(height-1, math.Floor(math.Max(v1.y, math.Max(v2.y, v3.y)))))

			// Algoritma yang digunakan untuk menentukan apakah suatu pixel berada di dalam segitiga atau tidak (algoritma barycentric).
			triangleArea := (v1.y-v3.y)*(v2.x-v3.x) + (v2.y-v3.y)*(v3.x-v1.x)

			// Untuk setiap pixel yang berada di dalam segitiga, kita akan menentukan kedalaman pixel tersebut dan mengganti warna pixel tersebut.
			for y := minY; y <= maxY; y++ {
				for x := minX; x <= maxX; x++ {
					// Menghitung koordinat barycentric.
					b1 := ((float64(y)-v3.y)*(v2.x-v3.x) + (v2.y-v3.y)*(v3.x-float64(x))) / triangleArea
					b2 := ((float64(y)-v1.y)*(v3.x-v1.x) + (v3.y-v1.y)*(v1.x-float64(x))) / triangleArea
					b3 := ((float64(y)-v2.y)*(v1.x-v2.x) + (v1.y-v2.y)*(v2.x-float64(x))) / triangleArea

					// Jika koordinat barycentric berada di dalam segitiga, maka kita akan mengganti warna pixel tersebut.
					if b1 >= 0 && b1 <= 1 && b2 >= 0 && b2 <= 1 && b3 >= 0 && b3 <= 1 {
						depth := b1*v1.z + b2*v2.z + b3*v3.z
						zIndex := y*int(width) + x

						// Jika kedalaman pixel tersebut lebih besar dari kedalaman pixel yang telah ada sebelumnya, maka kita akan mengganti warna pixel tersebut.
						// Warna pixel tersebut akan diubah berdasarkan sudut antara normal dan vektor pandang.
						if zBuffer[zIndex] < depth {
							img.Set(x, y, getShade(t.color, angleCos))
							zBuffer[zIndex] = depth
						}
					}
				}
			}
		}

		// Menambahkan gambar yang telah dibuat ke dalam animasi GIF.
		// Setiap gambar akan memiliki delay sebesar 3/100 detik.
		paletteImg := image.NewPaletted(img.Bounds(), color.Palette{colorBlack, colorWhite, colorRed, colorGreen, colorBlue})
		draw.Draw(paletteImg, img.Bounds(), img, image.Point{}, draw.Src)
		anim.Delay = append(anim.Delay, 3)
		anim.Image = append(anim.Image, paletteImg)
	}

	// Menyimpan animasi GIF ke dalam file.
	f, _ := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0600)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	err := gif.EncodeAll(f, &anim)
	if err != nil {
		return
	}
}

// generateSphere digunakan untuk membuat bola dunia.
// Bola dunia ini akan terdiri dari beberapa segitiga yang membentuk bola.
// Setiap segitiga akan diberi warna yang berbeda-beda.
// Pada parameter pertama dari fungsi generateSphere, kita bisa mengatur jumlah garis lintang dan bujur yang akan digunakan untuk membuat bola.
// Pada parameter kedua dari fungsi generateSphere, kita bisa mengatur jumlah garis bujur yang akan digunakan untuk membuat bola.
// Semakin besar nilai dari kedua parameter tersebut, maka bola yang dihasilkan akan semakin halus.
func generateSphere(latitudes, longitudes int) []Triangle {
	// Membuat array untuk menyimpan segitiga-segitiga yang akan membentuk bola.
	var tris []Triangle
	// Radius dari bola. Semakin besar nilai radius, maka bola yang dihasilkan akan semakin besar.
	radius := 100.0

	// Membuat segitiga-segitiga yang membentuk bola.
	for i := 0; i < latitudes; i++ {
		for j := 0; j < longitudes; j++ {
			// Menghitung titik-titik segitiga. Setiap segitiga akan terdiri dari 3 titik.
			theta1 := float64(i) * math.Pi / float64(latitudes)
			theta2 := float64(i+1) * math.Pi / float64(latitudes)
			phi1 := float64(j) * 2 * math.Pi / float64(longitudes)
			phi2 := float64(j+1) * 2 * math.Pi / float64(longitudes)

			v1 := Vertex{
				radius * math.Sin(theta1) * math.Cos(phi1),
				radius * math.Sin(theta1) * math.Sin(phi1),
				radius * math.Cos(theta1),
			}
			v2 := Vertex{
				radius * math.Sin(theta1) * math.Cos(phi2),
				radius * math.Sin(theta1) * math.Sin(phi2),
				radius * math.Cos(theta1),
			}
			v3 := Vertex{
				radius * math.Sin(theta2) * math.Cos(phi2),
				radius * math.Sin(theta2) * math.Sin(phi2),
				radius * math.Cos(theta2),
			}
			v4 := Vertex{
				radius * math.Sin(theta2) * math.Cos(phi1),
				radius * math.Sin(theta2) * math.Sin(phi1),
				radius * math.Cos(theta2),
			}

			// Menentukan warna segitiga. Setiap segitiga akan memiliki warna yang berbeda-beda.
			// Warna segitiga akan bergantian antara merah, hijau, biru, dan kuning.
			var clr color.RGBA
			region := (i/2 + j/2) % 4
			switch region {
			case 0:
				clr = color.RGBA{R: 255, A: 255} // Red
			case 1:
				clr = color.RGBA{G: 255, A: 255} // Green
			case 2:
				clr = color.RGBA{B: 255, A: 255} // Blue
			case 3:
				clr = color.RGBA{R: 255, G: 255, A: 255} // Yellow
			}

			// Menambahkan segitiga ke dalam array tris.
			// Setiap segitiga akan memiliki 3 titik dan 1 warna.
			// Titik-titik tersebut akan membentuk segitiga dengan urutan searah jarum jam.
			tris = append(tris, Triangle{v1, v2, v3, clr})
			tris = append(tris, Triangle{v1, v3, v4, clr})
		}
	}

	// Membuat segitiga-segitiga yang membentuk bola bagian atas dan bawah.
	for j := 0; j < longitudes; j++ {
		// Membuat segitiga bagian atas. Segitiga ini akan memiliki titik puncak di bagian atas bola.
		vTop := Vertex{0, 0, radius}
		v1 := Vertex{
			radius * math.Sin(math.Pi/float64(latitudes)) * math.Cos(float64(j)*2*math.Pi/float64(longitudes)),
			radius * math.Sin(math.Pi/float64(latitudes)) * math.Sin(float64(j)*2*math.Pi/float64(longitudes)),
			radius * math.Cos(math.Pi/float64(latitudes)),
		}
		v2 := Vertex{
			radius * math.Sin(math.Pi/float64(latitudes)) * math.Cos(float64((j+1)%longitudes)*2*math.Pi/float64(longitudes)),
			radius * math.Sin(math.Pi/float64(latitudes)) * math.Sin(float64((j+1)%longitudes)*2*math.Pi/float64(longitudes)),
			radius * math.Cos(math.Pi/float64(latitudes)),
		}

		// Menambahkan segitiga ke dalam array tris.
		tris = append(tris, Triangle{vTop, v1, v2, colorWhite})

		// Membuat segitiga bagian bawah. Segitiga ini akan memiliki titik puncak di bagian bawah bola.
		vBottom := Vertex{0, 0, -radius}
		v3 := Vertex{
			radius * math.Sin(-math.Pi/float64(latitudes)) * math.Cos(float64(j)*2*math.Pi/float64(longitudes)),
			radius * math.Sin(-math.Pi/float64(latitudes)) * math.Sin(float64(j)*2*math.Pi/float64(longitudes)),
			radius * math.Cos(-math.Pi/float64(latitudes)),
		}
		v4 := Vertex{
			radius * math.Sin(-math.Pi/float64(latitudes)) * math.Cos(float64((j+1)%longitudes)*2*math.Pi/float64(longitudes)),
			radius * math.Sin(-math.Pi/float64(latitudes)) * math.Sin(float64((j+1)%longitudes)*2*math.Pi/float64(longitudes)),
			radius * math.Cos(-math.Pi/float64(latitudes)),
		}

		// Menambahkan segitiga ke dalam array tris.
		tris = append(tris, Triangle{vBottom, v3, v4, colorWhite})
	}

	// Membuat sebuat stik di bagian atas dan bawah bola.
	// Stik ini akan membantu kita untuk mengetahui posisi bola.
	stickWidth := 2.0
	stickLength := 400.0

	for i := 0; i < longitudes; i++ {
		phi1 := float64(i) * 2 * math.Pi / float64(longitudes)
		phi2 := float64((i+1)%longitudes) * 2 * math.Pi / float64(longitudes)

		v1 := Vertex{
			stickWidth * math.Cos(phi1),
			stickWidth * math.Sin(phi1),
			-(radius + stickLength/2),
		}
		v2 := Vertex{
			stickWidth * math.Cos(phi1),
			stickWidth * math.Sin(phi1),
			-radius,
		}
		v3 := Vertex{
			stickWidth * math.Cos(phi2),
			stickWidth * math.Sin(phi2),
			-radius,
		}
		v4 := Vertex{
			stickWidth * math.Cos(phi2),
			stickWidth * math.Sin(phi2),
			-(radius + stickLength/2),
		}

		// Menambahkan stik ke dalam array tris.
		tris = append(tris, Triangle{v1, v2, v3, colorWhite})
		tris = append(tris, Triangle{v1, v3, v4, colorWhite})

		v1 = Vertex{
			stickWidth * math.Cos(phi1),
			stickWidth * math.Sin(phi1),
			radius + stickLength/2,
		}
		v2 = Vertex{
			stickWidth * math.Cos(phi1),
			stickWidth * math.Sin(phi1),
			radius,
		}
		v3 = Vertex{
			stickWidth * math.Cos(phi2),
			stickWidth * math.Sin(phi2),
			radius,
		}
		v4 = Vertex{
			stickWidth * math.Cos(phi2),
			stickWidth * math.Sin(phi2),
			radius + stickLength/2,
		}

		// Menambahkan stik ke dalam array tris.
		tris = append(tris, Triangle{v1, v2, v3, colorWhite})
		tris = append(tris, Triangle{v1, v3, v4, colorWhite})
	}

	return tris
}

// getShade digunakan untuk mengubah warna pixel berdasarkan sudut antara normal dan vektor pandang.
func getShade(c color.Color, shade float64) color.Color {
	r, g, b, _ := c.RGBA()

	redLinear := math.Pow(float64(r)/0xffff, 2.4) * shade
	greenLinear := math.Pow(float64(g)/0xffff, 2.4) * shade
	blueLinear := math.Pow(float64(b)/0xffff, 2.4) * shade

	red := uint8(math.Pow(redLinear, 1/2.4) * 0xff)
	green := uint8(math.Pow(greenLinear, 1/2.4) * 0xff)
	blue := uint8(math.Pow(blueLinear, 1/2.4) * 0xff)

	return color.RGBA{R: red, G: green, B: blue, A: 0xff}
}

// Matrix3 merepresentasikan matriks 3x3.
// Matriks ini digunakan untuk melakukan transformasi terhadap titik-titik dalam ruang 3D.
// Matriks ini juga digunakan untuk melakukan rotasi terhadap bola dunia.
type Matrix3 [3][3]float64

// multiply digunakan untuk mengalikan dua matriks 3x3.
// Hasil dari perkalian kedua matriks tersebut akan disimpan ke dalam matriks baru.
// Matriks baru tersebut akan digunakan untuk melakukan rotasi terhadap bola dunia.
func (m Matrix3) multiply(other Matrix3) Matrix3 {
	var result Matrix3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				result[i][j] += m[i][k] * other[k][j]
			}
		}
	}
	return result
}

// transform digunakan untuk melakukan transformasi terhadap titik-titik dalam ruang 3D.
// Hasil dari transformasi tersebut akan disimpan ke dalam titik baru.
// Titik baru tersebut akan digunakan untuk menentukan posisi segitiga pada gambar.
func (m Matrix3) transform(v Vertex) Vertex {
	x := m[0][0]*v.x + m[0][1]*v.y + m[0][2]*v.z
	y := m[1][0]*v.x + m[1][1]*v.y + m[1][2]*v.z
	z := m[2][0]*v.x + m[2][1]*v.y + m[2][2]*v.z
	return Vertex{x, y, z}
}
