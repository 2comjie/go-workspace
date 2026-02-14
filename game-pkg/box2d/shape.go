package box2d

import (
	"game-pkg/math/vec2"
	"math"
)

type ShapeType int

const (
	ShapeCircle ShapeType = iota
	ShapeRectangle
)

type Shape interface {
	Type() ShapeType
	GetAABB(pos vec2.Vector2) AABB
}

type AABB struct {
	Min vec2.Vector2
	Max vec2.Vector2
}

func NewAABB(min, max vec2.Vector2) AABB {
	return AABB{
		Min: min,
		Max: max,
	}
}

func (a AABB) Contains(point vec2.Vector2) bool {
	return point.X >= a.Min.X && point.X <= a.Max.X &&
		point.Y >= a.Min.Y && point.Y <= a.Max.Y
}

func (a AABB) Intersects(other AABB) bool {
	return a.Min.X <= other.Max.X && a.Max.X >= other.Min.X &&
		a.Min.Y <= other.Max.Y && a.Max.Y >= other.Min.Y
}

func (a AABB) Center() vec2.Vector2 {
	return vec2.New(
		(a.Min.X+a.Max.X)/2,
		(a.Min.Y+a.Max.Y)/2,
	)
}

func (a AABB) Size() vec2.Vector2 {
	return vec2.New(
		a.Max.X-a.Min.X,
		a.Max.Y-a.Min.Y,
	)
}

type Circle struct {
	Radius float64
}

func NewCircle(radius float64) *Circle {
	return &Circle{
		Radius: radius,
	}
}

func (c *Circle) Type() ShapeType {
	return ShapeCircle
}

func (c *Circle) GetAABB(pos vec2.Vector2) AABB {
	return AABB{
		Min: vec2.New(pos.X-c.Radius, pos.Y-c.Radius),
		Max: vec2.New(pos.X+c.Radius, pos.Y+c.Radius),
	}
}

type Rectangle struct {
	Width  float64
	Height float64
}

func NewRectangle(width, height float64) *Rectangle {
	return &Rectangle{Width: width, Height: height}
}

func (r *Rectangle) Type() ShapeType {
	return ShapeRectangle
}

func (r *Rectangle) GetAABB(position vec2.Vector2) AABB {
	halfW := r.Width / 2
	halfH := r.Height / 2
	return AABB{
		Min: vec2.New(position.X-halfW, position.Y-halfH),
		Max: vec2.New(position.X+halfW, position.Y+halfH),
	}
}

type CollisionInfo struct {
	Colliding bool         // 是否碰撞
	Normal    vec2.Vector2 // 碰撞法线
	Depth     float64      // 穿透深度
}

func CheckCollision(shapeA Shape, posA vec2.Vector2, shapeB Shape, posB vec2.Vector2) CollisionInfo {
	// AABB粗检测
	aabbA := shapeA.GetAABB(posA)
	aabbB := shapeB.GetAABB(posB)

	if !aabbA.Intersects(aabbB) {
		return CollisionInfo{Colliding: false}
	}

	// 精确碰撞检测
	typeA := shapeA.Type()
	typeB := shapeB.Type()

	if typeA == ShapeCircle && typeB == ShapeCircle {
		return checkCircleCircle(shapeA.(*Circle), posA, shapeB.(*Circle), posB)
	} else if typeA == ShapeRectangle && typeB == ShapeRectangle {
		return checkRectRect(shapeA.(*Rectangle), posA, shapeB.(*Rectangle), posB)
	} else if typeA == ShapeCircle && typeB == ShapeRectangle {
		return checkCircleRect(shapeA.(*Circle), posA, shapeB.(*Rectangle), posB)
	} else if typeA == ShapeRectangle && typeB == ShapeCircle {
		info := checkCircleRect(shapeB.(*Circle), posB, shapeA.(*Rectangle), posA)
		info.Normal = info.Normal.Neg() // 反转法线方向
		return info
	}

	return CollisionInfo{Colliding: false}
}

func checkCircleCircle(a *Circle, posA vec2.Vector2, b *Circle, posB vec2.Vector2) CollisionInfo {
	delta := posB.Sub(posA)
	distance := delta.Len()
	radiusSum := a.Radius + b.Radius

	if distance >= radiusSum {
		return CollisionInfo{Colliding: false}
	}

	if distance == 0 {
		// 完全重叠，返回默认方向
		return CollisionInfo{
			Colliding: true,
			Normal:    vec2.New(1, 0),
			Depth:     radiusSum,
		}
	}

	return CollisionInfo{
		Colliding: true,
		Normal:    delta.Div(distance), // 归一化
		Depth:     radiusSum - distance,
	}
}

func checkRectRect(a *Rectangle, posA vec2.Vector2, b *Rectangle, posB vec2.Vector2) CollisionInfo {
	aabbA := a.GetAABB(posA)
	aabbB := b.GetAABB(posB)

	// 计算重叠
	overlapX := math.Min(aabbA.Max.X, aabbB.Max.X) - math.Max(aabbA.Min.X, aabbB.Min.X)
	overlapY := math.Min(aabbA.Max.Y, aabbB.Max.Y) - math.Max(aabbA.Min.Y, aabbB.Min.Y)

	if overlapX <= 0 || overlapY <= 0 {
		return CollisionInfo{Colliding: false}
	}

	// 选择最小穿透轴
	var normal vec2.Vector2
	var depth float64

	if overlapX < overlapY {
		depth = overlapX
		if posA.X < posB.X {
			normal = vec2.New(-1, 0)
		} else {
			normal = vec2.New(1, 0)
		}
	} else {
		depth = overlapY
		if posA.Y < posB.Y {
			normal = vec2.New(0, -1)
		} else {
			normal = vec2.New(0, 1)
		}
	}

	return CollisionInfo{
		Colliding: true,
		Normal:    normal,
		Depth:     depth,
	}
}

func checkCircleRect(circle *Circle, circlePos vec2.Vector2, rect *Rectangle, rectPos vec2.Vector2) CollisionInfo {
	aabb := rect.GetAABB(rectPos)

	// 找到矩形上最接近圆心的点
	closestX := math.Max(aabb.Min.X, math.Min(circlePos.X, aabb.Max.X))
	closestY := math.Max(aabb.Min.Y, math.Min(circlePos.Y, aabb.Max.Y))
	closest := vec2.New(closestX, closestY)

	// 计算圆心到最近点的距离
	delta := circlePos.Sub(closest)
	distance := delta.Len()

	if distance >= circle.Radius {
		return CollisionInfo{Colliding: false}
	}

	// 圆心在矩形内部
	if distance == 0 {
		// 计算最近边界
		distToLeft := circlePos.X - aabb.Min.X
		distToRight := aabb.Max.X - circlePos.X
		distToBottom := circlePos.Y - aabb.Min.Y
		distToTop := aabb.Max.Y - circlePos.Y

		minDist := math.Min(math.Min(distToLeft, distToRight), math.Min(distToBottom, distToTop))

		if minDist == distToLeft {
			return CollisionInfo{Colliding: true, Normal: vec2.New(-1, 0), Depth: circle.Radius + distToLeft}
		} else if minDist == distToRight {
			return CollisionInfo{Colliding: true, Normal: vec2.New(1, 0), Depth: circle.Radius + distToRight}
		} else if minDist == distToBottom {
			return CollisionInfo{Colliding: true, Normal: vec2.New(0, -1), Depth: circle.Radius + distToBottom}
		} else {
			return CollisionInfo{Colliding: true, Normal: vec2.New(0, 1), Depth: circle.Radius + distToTop}
		}
	}

	return CollisionInfo{
		Colliding: true,
		Normal:    delta.Div(distance), // 归一化
		Depth:     circle.Radius - distance,
	}
}
