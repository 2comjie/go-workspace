package box2d

import "game-pkg/math/vec2"

const (
	// 每个节点最多容纳的对象数
	MaxObjectsPerNode = 8
	// 最大深度
	MaxDepth = 6
)

type QuadTree struct {
	level    int
	bounds   AABB
	objects  []*Collider
	children [4]*QuadTree // 四个子节点: [0]左上 [1]右上 [2]左下 [3]右下
}

func NewQuadTree(level int, bounds AABB) *QuadTree {
	return &QuadTree{
		level:   level,
		bounds:  bounds,
		objects: make([]*Collider, 0, MaxObjectsPerNode),
	}
}

func (qt *QuadTree) Clear() {
	qt.objects = qt.objects[:0]
	for i := range qt.children {
		if qt.children[i] != nil {
			qt.children[i].Clear()
			qt.children[i] = nil
		}
	}
}

func (qt *QuadTree) Split() {
	subWidth := (qt.bounds.Max.X - qt.bounds.Min.X) / 2
	subHeight := (qt.bounds.Max.Y - qt.bounds.Min.Y) / 2

	x := qt.bounds.Min.X
	y := qt.bounds.Min.Y

	// 左上
	qt.children[0] = NewQuadTree(qt.level+1, NewAABB(
		vec2.New(x, y+subHeight),
		vec2.New(x+subWidth, y+subHeight*2),
	))
	// 右上
	qt.children[1] = NewQuadTree(qt.level+1, NewAABB(
		vec2.New(x+subWidth, y+subHeight),
		vec2.New(x+subWidth*2, y+subHeight*2),
	))
	// 左下
	qt.children[2] = NewQuadTree(qt.level+1, NewAABB(
		vec2.New(x, y),
		vec2.New(x+subWidth, y+subHeight),
	))
	// 右下
	qt.children[3] = NewQuadTree(qt.level+1, NewAABB(
		vec2.New(x+subWidth, y),
		vec2.New(x+subWidth*2, y+subHeight),
	))
}

func (qt *QuadTree) GetIndex(aabb AABB) int {
	index := -1
	midX := (qt.bounds.Min.X + qt.bounds.Max.X) / 2
	midY := (qt.bounds.Min.Y + qt.bounds.Max.Y) / 2

	// 判断是否在上半部分
	topHalf := aabb.Min.Y >= midY
	// 判断是否在下半部分
	bottomHalf := aabb.Max.Y <= midY

	// 判断是否在左半部分
	if aabb.Max.X <= midX {
		if topHalf {
			index = 0 // 左上
		} else if bottomHalf {
			index = 2 // 左下
		}
	} else if aabb.Min.X >= midX { // 右半部分
		if topHalf {
			index = 1 // 右上
		} else if bottomHalf {
			index = 3 // 右下
		}
	}
	return index
}

func (qt *QuadTree) Insert(collider *Collider) {
	if collider == nil || collider.Body == nil {
		return
	}

	// 如果已经分割，尝试插入子节点
	if qt.children[0] != nil {
		aabb := collider.GetAABB()
		index := qt.GetIndex(aabb)
		if index != -1 {
			qt.children[index].Insert(collider)
			return
		}
	}

	// 没有分割，插入这个节点
	qt.objects = append(qt.objects, collider)

	// 检查是否需要分割
	if len(qt.objects) > MaxObjectsPerNode && qt.level < MaxDepth {
		// 如果还没分割，进行分割
		if qt.children[0] == nil {
			qt.Split()
		}

		// 尝试将对象移动到子节点
		i := 0
		for i < len(qt.objects) {
			aabb := qt.objects[i].GetAABB()
			index := qt.GetIndex(aabb)
			if index != -1 {
				qt.children[index].Insert(qt.objects[i])
				// 从当前节点移除
				qt.objects[i] = qt.objects[len(qt.objects)-1]
				qt.objects = qt.objects[:len(qt.objects)-1]
			} else {
				i++
			}
		}
	}
}
