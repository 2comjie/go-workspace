package base_game

import (
	"hutool/reflectx"
	"reflect"
	"time"
)

type EntityId int64

type IEntity interface {
	OnStart(world *World)
	OnUpdate(world *World, dt time.Duration)
	OnDestroy(world *World)
}

type World struct {
	idToEntity   map[EntityId]IEntity
	typeToEntity map[reflect.Type]map[EntityId]IEntity
	nextId       EntityId
}

func NewWorld() *World {
	return &World{
		idToEntity:   make(map[EntityId]IEntity),
		typeToEntity: make(map[reflect.Type]map[EntityId]IEntity),
		nextId:       0,
	}
}

func GenerateEntity[T IEntity](world *World) T {
	entity := reflectx.NewPointerIns2[T]()
	id := world.nextId
	world.nextId++
	world.idToEntity[id] = entity
	rType := reflectx.TypeOf(entity)
	if world.typeToEntity[rType] == nil {
		world.typeToEntity[rType] = make(map[EntityId]IEntity)
	}
	world.typeToEntity[rType][id] = entity
	return entity
}

func GetEntityList[T IEntity](world *World) []T {
	rType := reflectx.GenericTypeOf[T]()
	typeToEntity, ok := world.typeToEntity[rType]
	if !ok {
		return nil
	}
	outList := make([]T, len(typeToEntity))
	for _, v := range typeToEntity {
		outList = append(outList, v.(T))
	}
	return outList
}

func GetEntity[T IEntity](world *World) T {
	var zero T
	rType := reflectx.GenericTypeOf[T]()
	typeToEntity, ok := world.typeToEntity[rType]
	if !ok {
		return zero
	}
	for _, v := range typeToEntity {
		return v.(T)
	}
	return zero
}

func (w *World) Destroy(entity IEntity) {
	entityId, ok := w.getEntityId(entity)
	if !ok {
		panic("entity not found")
	}
	entity.OnDestroy(w)
	delete(w.idToEntity, entityId)
	rType := reflectx.TypeOf(entity)
	delete(w.typeToEntity[rType], entityId)
	if len(w.typeToEntity[rType]) == 0 {
		delete(w.typeToEntity, rType)
	}
}

func (w *World) getEntityId(entity IEntity) (EntityId, bool) {
	for id, v := range w.idToEntity {
		if v == entity {
			return id, true
		}
	}
	return 0, false
}
