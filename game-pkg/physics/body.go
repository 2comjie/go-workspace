package physics

import (
	"game-pkg/vector2"
	"math"
)

type ShapeType int32

const (
	Circle    ShapeType = 0
	Rectangle           = 1
)

type Body struct {
	pos              vector2.Vector2
	linearVelocity   vector2.Vector2
	rotation         float64
	rotationVelocity float64
	density          float64
	mass             float64
	area             float64
	isStatic         bool

	radius    float64
	width     float64
	height    float64
	shapeType ShapeType
}

func CircleBody(radius float64, pos vector2.Vector2, density float64, isStatic bool, restitution float64) *Body {
	area := math.Pi * radius * radius
	mass := area * density
	return &Body{
		pos:              pos,
		linearVelocity:   vector2.Vector2{},
		rotation:         0,
		rotationVelocity: 0,
		density:          density,
		mass:             mass,
		area:             area,
		isStatic:         isStatic,
		radius:           radius,
		width:            0,
		height:           0,
		shapeType:        Circle,
	}
}

func RectangleBody(width, height float64, pos vector2.Vector2, density float64, isStatic bool, restitution float64) *Body {
	area := width * height
	mass := area * density
	return &Body{
		pos:              pos,
		linearVelocity:   vector2.Vector2{},
		rotation:         0,
		rotationVelocity: 0,
		density:          density,
		mass:             mass,
		area:             area,
		isStatic:         isStatic,
		radius:           0,
		width:            width,
		height:           height,
		shapeType:        Rectangle,
	}
}
