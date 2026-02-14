package box2d

import "game-pkg/math/vec2"

type Layer uint32

const (
	LayerDefault Layer = 1 << 0     // 默认层
	Layer1       Layer = 1 << 1     // 自定义层1
	Layer2       Layer = 1 << 2     // 自定义层2
	Layer3       Layer = 1 << 3     // 自定义层3
	Layer4       Layer = 1 << 4     // 自定义层4
	Layer5       Layer = 1 << 5     // 自定义层5
	Layer6       Layer = 1 << 6     // 自定义层6
	Layer7       Layer = 1 << 7     // 自定义层7
	LayerAll     Layer = 0xFFFFFFFF // 所有层
)

func CanCollide(layerA, maskA Layer, layerB, maskB Layer) bool {
	return (layerA&maskB) != 0 && (layerB&maskA) != 0
}

type CollisionHandler interface {
	OnCollisionEnter(other *Collider)
	OnCollisionExit(other *Collider)
	OnCollisionStay(other *Collider)
}

type Collider struct {
	Body              *RigidBody
	Shape             Shape
	Layer             Layer // 当前对象所在层级
	Mask              Layer // 可以和哪些层级碰撞
	IsTrigger         bool  // 是否是触发器 不产生物理效果
	Handler           CollisionHandler
	AmountedComponent any // 挂载的对象
	collidingWith     map[*Collider]bool
}

func (c *Collider) GetAABB() AABB {
	if c.Body == nil || c.Shape == nil {
		return AABB{}
	}
	return c.Shape.GetAABB(c.Body.Pos)
}

func (c *Collider) CheckCollision(other *Collider) CollisionInfo {
	if c.Body == nil || c.Shape == nil || other.Body == nil || other.Shape == nil {
		return CollisionInfo{Colliding: false}
	}
	// 检查层级
	if !CanCollide(c.Layer, c.Mask, other.Layer, other.Mask) {
		return CollisionInfo{Colliding: false}
	}

	return CheckCollision(c.Shape, c.Body.Pos, other.Shape, other.Body.Pos)
}

// ResolveCollision 解决碰撞响应
func (c *Collider) ResolveCollision(other *Collider, info CollisionInfo) {
	// 触发器不产生物理响应
	if c.IsTrigger || other.IsTrigger {
		return
	}

	// 至少有一个是动态刚体才需要响应
	if c.Body.Type == Static && other.Body.Type == Static {
		return
	}

	// 分离对象
	invMassA := c.Body.GetInverseMass()
	invMassB := other.Body.GetInverseMass()
	totalInvMass := invMassA + invMassB

	if totalInvMass == 0 {
		return
	}

	// 位置修正
	correction := info.Normal.Mul(info.Depth / totalInvMass)
	if c.Body.Type == Dynamic {
		c.Body.Pos = c.Body.Pos.Sub(correction.Mul(invMassA))
	}
	if other.Body.Type == Dynamic {
		other.Body.Pos = other.Body.Pos.Add(correction.Mul(invMassB))
	}

	// 速度响应 (考虑弹性和摩擦)
	relativeVel := other.Body.Velocity.Sub(c.Body.Velocity)
	velAlongNormal := relativeVel.Dot(info.Normal)

	// 物体正在分离，不需要响应
	if velAlongNormal > 0 {
		return
	}

	// 计算反弹系数
	restitution := (c.Body.Restitution + other.Body.Restitution) / 2

	// 计算冲量标量
	impulseScalar := -(1 + restitution) * velAlongNormal
	impulseScalar /= totalInvMass

	// 应用冲量
	impulse := info.Normal.Mul(impulseScalar)
	if c.Body.Type == Dynamic {
		c.Body.ApplyImpulse(impulse.Neg())
	}
	if other.Body.Type == Dynamic {
		other.Body.ApplyImpulse(impulse)
	}

	// 应用摩擦
	applyFriction(c, other, info, relativeVel, totalInvMass)
}

func applyFriction(a, b *Collider, info CollisionInfo, relativeVel vec2.Vector2, totalInvMass float64) {
	// 计算切线方向
	tangent := relativeVel.Sub(info.Normal.Mul(relativeVel.Dot(info.Normal)))
	tangentLen := tangent.Len()
	if tangentLen < 0.001 {
		return
	}
	tangent = tangent.Div(tangentLen) // 归一化

	// 计算摩擦系数 (取平均值)
	friction := (a.Body.Friction + b.Body.Friction) / 2

	// 计算摩擦冲量
	frictionImpulse := -tangent.Dot(relativeVel) * friction
	frictionImpulse /= totalInvMass

	// 应用摩擦冲量
	frictionVec := tangent.Mul(frictionImpulse)
	if a.Body.Type == Dynamic {
		a.Body.ApplyImpulse(frictionVec.Neg())
	}
	if b.Body.Type == Dynamic {
		b.Body.ApplyImpulse(frictionVec)
	}
}

func (c *Collider) HandleCollisionCallbacks(other *Collider, isColliding bool) {
	if c.Handler == nil {
		return
	}

	wasColliding := c.collidingWith[other]

	if isColliding {
		if !wasColliding {
			// 新碰撞
			c.collidingWith[other] = true
			c.Handler.OnCollisionEnter(other)
		} else {
			// 持续碰撞
			c.Handler.OnCollisionStay(other)
		}
	} else {
		if wasColliding {
			// 碰撞结束
			delete(c.collidingWith, other)
			c.Handler.OnCollisionExit(other)
		}
	}
}

func (c *Collider) ClearCollisionState() {
	c.collidingWith = make(map[*Collider]bool)
}
