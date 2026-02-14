package vector2

import (
	"fmt"
	"math"
)

type Vector2 struct {
	X float64
	Y float64
}

func New(x, y float64) Vector2 {
	return Vector2{X: x, Y: y}
}

func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vector2) Sub(other Vector2) Vector2 {
	return Vector2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vector2) Mul(scaler float64) Vector2 {
	return Vector2{X: v.X * scaler, Y: v.Y * scaler}
}

func (v Vector2) Div(scaler float64) Vector2 {
	return Vector2{X: v.X / scaler, Y: v.Y / scaler}
}

func (v Vector2) Neg() Vector2 {
	return Vector2{X: -v.X, Y: -v.Y}
}

func Zero() Vector2 {
	return Vector2{X: 0, Y: 0}
}

func (v Vector2) Equal(other Vector2) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vector2) String() string {
	return fmt.Sprintf("(%f, %f)", v.X, v.Y)
}

func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2) Distance(other Vector2) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vector2) Normalize() Vector2 {
	l := v.Length()
	return Vector2{X: v.X / l, Y: v.Y / l}
}

func (v Vector2) Dot(other Vector2) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vector2) Cross(other Vector2) float64 {
	return v.X*other.Y - v.Y*other.X
}
