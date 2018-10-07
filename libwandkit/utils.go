package libwandkit

import "math"

// QuantarionTo2d struct to deal with Quantarion conversion to 2D
type QuantarionTo2d struct {
	quantarion Quantarion
	canvasX    int
	canvasY    int
	matrix     [16]float64
	zeroVector Vector
	oneVector  Vector
	euler      Vector
}

// Quantarion struct ...
type Quantarion struct {
	x float64
	y float64
	z float64
	w float64
}

// Vector struct ...
type Vector struct {
	x float64
	y float64
	z float64
}

// NewQuantarionTo2d return new instance of QuantarionTo2d
func NewQuantarionTo2d(x, y int, quantarion Quantarion) *QuantarionTo2d {
	return &QuantarionTo2d{
		quantarion: quantarion,
		canvasX:    x,
		canvasY:    y,
		matrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
		zeroVector: Vector{x: 0, y: 0, z: 0},
		oneVector:  Vector{x: 1, y: 1, z: 1},
		euler:      Vector{x: 0, y: 0, z: 0},
	}
}

// Position turns quantarion to 2d geometry
func (q *QuantarionTo2d) Position() (x, y, pitch, roll, yaw float64) {
	quat := Quantarion{}
	quat.x = q.quantarion.x / 1024
	quat.y = q.quantarion.y / 1024
	quat.z = q.quantarion.z / 1024
	quat.w = q.quantarion.w / 1024
	norm := q.quantarion
	normLen := math.Sqrt(math.Pow(quat.y, 2) + math.Pow(quat.z, 2) + math.Pow(quat.w, 2) + math.Pow(quat.x, 2))
	if normLen == 0 {
		norm = Quantarion{x: 0, y: 0, z: 0, w: 1}
	} else {
		normLen = 1 / normLen
		norm.x = normLen
		norm.y = normLen
		norm.z = normLen
		norm.w = normLen
	}
	position := q.zeroVector
	x = norm.x
	y = norm.y
	z := norm.z
	w := norm.w
	scale := q.oneVector
	te := q.matrix

	xx := x * (x + x)
	xy := x * (y + y)
	xz := x * (z + z)
	yy := y * (y + y)
	yz := y * (z + z)
	zz := z * (z + z)
	wx := w * (x + x)
	wy := w * (y + y)
	wz := w * (z + z)
	sx := float64(scale.x)
	sy := float64(scale.y)
	sz := float64(scale.z)
	te[0] = (1.0 - (yy + zz)) * sx
	te[1] = (xy + wz) * sx
	te[2] = (xz - wy) * sx
	te[3] = 0
	te[4] = (xy - wz) * sy
	te[5] = (1 - (xx + zz)) * sy
	te[6] = (yz + wx) * sy
	te[7] = 0
	te[8] = (xz + wy) * sz
	te[9] = (yz - wx) * sz
	te[10] = (1 - (xx + yy)) * sz
	te[11] = 0
	te[12] = position.x
	te[13] = position.y
	te[14] = position.z
	te[15] = 1

	matrix11 := te[0]
	matrix12 := te[4]
	matrix13 := te[8]
	matrix22 := te[5]
	matrix23 := te[9]
	matrix32 := te[6]
	matrix33 := te[10]
	y = math.Asin(math.Max(-1, math.Min(1, matrix13)))
	if math.Abs(matrix13) < 0.99999 {
		x = math.Atan2(-matrix23, matrix33)
		z = math.Atan2(-matrix12, matrix11)
	} else {
		x = math.Atan2(matrix32, matrix22)
		z = 0
	}
	yawComplete := [3]float64{y, z, x * 2}
	for idx, angle := range yawComplete {
		yawComplete[idx] = angle * (180 / math.Pi)
		// yawComplete[idx] = real(yawComplete[idx])
	}
	yaw = yawComplete[0]
	roll = yawComplete[1]
	pitch = yawComplete[2]

	w = float64(q.canvasX / 2)
	h := float64(q.canvasY / 2)
	x = -(((yaw / 180) * (w * 4)) - w)
	y = h - (3 * pitch)
	pitch = pitch / 2
	roll = -roll
	return
}
