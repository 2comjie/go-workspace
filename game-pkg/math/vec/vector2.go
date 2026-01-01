package vec

import "math"

type Vector2 struct {
	X float64
	Y float64
}

func (v Vector2) Add(other Vector2) Vector2 {
	r := Vector2{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
	return r
}

func (v Vector2) Sub(other Vector2) Vector2 {
	r := Vector2{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
	return r
}

func (v Vector2) Div(scalar float64) Vector2 {
	return Vector2{
		X: v.X / scalar,
		Y: v.Y / scalar,
	}
}

func (v Vector2) Dot(other Vector2) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vector2) Cross(other Vector2) float64 {
	return v.X*other.Y - v.Y*other.X
}

func (v Vector2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2) Dist(other Vector2) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vector2) Normalize() Vector2 {
	l := v.Len()
	if l == 0 {
		return Zero()
	}
	return v.Div(l)
}

func (v Vector2) Neg() Vector2 {
	return Vector2{
		X: -v.X,
		Y: -v.Y,
	}
}

func Zero() Vector2 {
	return Vector2{0, 0}
}

func UnitX() Vector2 {
	return Vector2{1, 0}
}

func UnitY() Vector2 {
	return Vector2{0, 1}
}

func Unit() Vector2 {
	return Vector2{1, 1}
}

func FromAngle(angle float64) Vector2 {
	return Vector2{
		X: math.Cos(angle),
		Y: math.Sin(angle),
	}
}

func FromAngleDeg(angleDeg float64) Vector2 {
	return Vector2{
		X: math.Cos(angleDeg * math.Pi / 180),
		Y: math.Sin(angleDeg * math.Pi / 180),
	}
}

func New(x, y float64) Vector2 {
	return Vector2{x, y}
}
