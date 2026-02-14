package box2d

import (
	"game-pkg/math/vec2"
	"time"
)

// 刚体类型
type BodyType int

const (
	// 静态刚体(不受力影响，不会移动)
	Static BodyType = iota
	// 动态刚体(受力影响)
	Dynamic
	// 运动学刚体(不受力影响，但可以设置速度)
	Kinematic
)

type RigidBody struct {
	Type        BodyType
	Pos         vec2.Vector2
	Velocity    vec2.Vector2
	Mass        float64
	Friction    float64 // 摩擦系数
	Restitution float64 // 弹性系数
	Drag        float64 // 空气阻力
	force       vec2.Vector2
	UseGravity  bool // 重力
}

func NewRigidBody(bodyType BodyType) *RigidBody {
	return &RigidBody{
		Type:        bodyType,
		Pos:         vec2.Zero(),
		Velocity:    vec2.Zero(),
		Mass:        1.0,
		Friction:    0.3,
		Restitution: 0.0,
		Drag:        0.01,
		force:       vec2.Zero(),
		UseGravity:  true,
	}
}

func (rb *RigidBody) AddForce(force vec2.Vector2) {
	if rb.Type == Dynamic {
		rb.force = rb.force.Add(force)
	}
}

func (rb *RigidBody) SetVelocity(velocity vec2.Vector2) {
	if rb.Type != Static {
		rb.Velocity = velocity
	}
}

// 应用冲量 弹性碰撞
func (rb *RigidBody) ApplyImpulse(impulse vec2.Vector2) {
	if rb.Type == Dynamic && rb.Mass > 0 {
		// v = v + impulse / mass
		rb.Velocity = rb.Velocity.Add(impulse.Div(rb.Mass))
	}
}

func (rb *RigidBody) Update(dt time.Duration, gravity vec2.Vector2) {
	if rb.Type == Static {
		return
	}

	if rb.Type == Dynamic {
		// 应用重力
		if rb.UseGravity {
			rb.AddForce(gravity.Mul(rb.Mass))
		}

		// 速度 v = v + a * dt
		if rb.Mass > 0 {
			// a = F / M
			acceleration := rb.force.Div(rb.Mass)
			rb.Velocity = rb.Velocity.Add(acceleration.Mul(dt.Seconds()))
		}
		rb.force = vec2.Zero()
		dragForce := 1.0 - rb.Drag
		if dragForce < 0 {
			dragForce = 0
		}
		rb.Velocity = rb.Velocity.Mul(dragForce) // 摩擦系数
	}
	rb.Pos = rb.Pos.Add(rb.Velocity.Mul(dt.Seconds()))
}

func (rb *RigidBody) GetInverseMass() float64 {
	if rb.Type == Static || rb.Mass == 0 {
		return 0
	}
	return 1.0 / rb.Mass
}
