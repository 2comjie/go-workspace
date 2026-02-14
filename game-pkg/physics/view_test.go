package physics

import (
	_ "embed"
	"game-pkg/vector2"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed view_test.html
var htmlContent string

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SimpleWorld 简单的物理世界，用于管理多个 Body
type SimpleWorld struct {
	bodies []*Body
	mu     sync.RWMutex
}

func NewSimpleWorld() *SimpleWorld {
	return &SimpleWorld{
		bodies: make([]*Body, 0),
	}
}

func (w *SimpleWorld) AddBody(body *Body) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bodies = append(w.bodies, body)
}

func (w *SimpleWorld) GetBodies() []*Body {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]*Body, len(w.bodies))
	copy(result, w.bodies)
	return result
}

// BodyData 用于 JSON 序列化的 Body 数据
type BodyData struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	ShapeType int32   `json:"shapeType"`
	Radius    float64 `json:"radius,omitempty"`
	Width     float64 `json:"width,omitempty"`
	Height    float64 `json:"height,omitempty"`
	Rotation  float64 `json:"rotation"`
}

// ClientMessage 客户端发送的消息
type ClientMessage struct {
	Type   string  `json:"type"` // "click"
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Button int     `json:"button"` // 0=左键, 2=右键
}

// ServerMessage 服务器发送的消息
type ServerMessage struct {
	Type   string     `json:"type"` // "update"
	Bodies []BodyData `json:"bodies"`
}

func TestPhysicsView(t *testing.T) {
	world := NewSimpleWorld()

	// 添加一些初始的静态地面
	ground1 := RectangleBody(800, 20, vector2.New(400, 590), 1.0, true, 0.5)
	world.AddBody(ground1)
	ground2 := RectangleBody(800, 20, vector2.New(400, 10), 1.0, true, 0.5)
	world.AddBody(ground2)
	wall1 := RectangleBody(20, 600, vector2.New(10, 300), 1.0, true, 0.5)
	world.AddBody(wall1)
	wall2 := RectangleBody(20, 600, vector2.New(790, 300), 1.0, true, 0.5)
	world.AddBody(wall2)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlContent))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// 发送初始状态
		sendUpdate(conn, world)

		// 启动定时更新
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()

		// 处理客户端消息
		go func() {
			for {
				var msg ClientMessage
				err := conn.ReadJSON(&msg)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket error: %v", err)
					}
					return
				}

				if msg.Type == "click" {
					pos := vector2.New(msg.X, msg.Y)
					if msg.Button == 0 { // 左键 - 生成圆
						body := CircleBody(20, pos, 1.0, false, 0.5)
						world.AddBody(body)
					} else if msg.Button == 2 { // 右键 - 生成方块
						body := RectangleBody(40, 40, pos, 1.0, false, 0.5)
						world.AddBody(body)
					}
				}
			}
		}()

		// 定时发送更新
		for range ticker.C {
			err := sendUpdate(conn, world)
			if err != nil {
				log.Printf("Send error: %v", err)
				return
			}
		}
	})

	log.Println("Physics visualization server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendUpdate(conn *websocket.Conn, world *SimpleWorld) error {
	bodies := world.GetBodies()
	bodyData := make([]BodyData, 0, len(bodies))

	for _, body := range bodies {
		bd := BodyData{
			X:         body.pos.X,
			Y:         body.pos.Y,
			ShapeType: int32(body.shapeType),
			Rotation:  body.rotation,
		}
		if body.shapeType == Circle {
			bd.Radius = body.radius
		} else if body.shapeType == Rectangle {
			bd.Width = body.width
			bd.Height = body.height
		}
		bodyData = append(bodyData, bd)
	}

	msg := ServerMessage{
		Type:   "update",
		Bodies: bodyData,
	}

	return conn.WriteJSON(msg)
}
