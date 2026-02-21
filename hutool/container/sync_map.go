package container

import (
	"sync"

	"golang.org/x/sync/singleflight"
)

type IMap[K comparable, V any] interface {
	Load(key K) (V, bool)
	Store(key K, v V)
	LoadOrStore(key K, v V) (V, bool)
	LoadOrStoreNew(key K, ctor func() V, release func(V)) (V, bool)
	LoadOrStoreNewWithSingleFlight(sfKey string, key K, ctor func() V, release func(V)) (V, bool)
	LoadAndDelete(key K) (V, bool)
	Delete(key K)
	Range(yield func(key K, v V) bool)
	DeleteIf(cond func(key K, v V) bool)
}

type SyncMap[K comparable, V any] struct {
	it   sync.Map
	nilV V
	sf   singleflight.Group
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	var nilV V
	return &SyncMap[K, V]{
		it:   sync.Map{},
		nilV: nilV,
		sf:   singleflight.Group{},
	}
}

func (s *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := s.it.Load(key)
	if !ok {
		return s.nilV, false
	}
	return v.(V), true
}

func (s *SyncMap[K, V]) Store(key K, v V) {
	s.it.Store(key, v)
}

func (s *SyncMap[K, V]) LoadOrStore(key K, v V) (V, bool) {
	actual, ok := s.it.LoadOrStore(key, v)
	return actual.(V), ok
}

func (s *SyncMap[K, V]) LoadOrStoreNew(key K, ctor func() V, release func(V)) (V, bool) {
	oldV, ok := s.it.Load(key)
	if ok {
		return oldV.(V), true
	}
	newV := ctor()
	tmpV, tmpOk := s.LoadOrStore(key, newV)
	if tmpOk && release != nil {
		release(newV)
	}
	return tmpV, tmpOk
}

type mapSfResult[V any] struct {
	loaded bool
	actual V
}

func (s *SyncMap[K, V]) LoadOrStoreNewWithSingleFlight(sfKey string, key K, ctor func() V, release func(V)) (V, bool) {
	sfRes, _, _ := s.sf.Do(sfKey, func() (interface{}, error) {
		// 加载旧的数据
		oldV, ok := s.it.Load(key)
		if ok {
			return mapSfResult[V]{loaded: true, actual: oldV.(V)}, nil
		}

		// 新的数据
		newV := ctor()
		tmpV, tmpOk := s.LoadOrStore(key, newV)
		if tmpOk && release != nil {
			release(newV)
		}

		return mapSfResult[V]{loaded: tmpOk, actual: tmpV}, nil
	})
	r := sfRes.(mapSfResult[V])
	return r.actual, r.loaded
}

func (s *SyncMap[K, V]) LoadAndDelete(key K) (V, bool) {
	v, ok := s.it.LoadAndDelete(key)
	if !ok {
		return s.nilV, false
	}
	return v.(V), true
}

func (s *SyncMap[K, V]) Delete(key K) {
	s.it.Delete(key)
}

func (s *SyncMap[K, V]) Range(yield func(key K, v V) bool) {
	s.it.Range(func(key, value any) bool {
		return yield(key.(K), value.(V))
	})
}

func (s *SyncMap[K, V]) DeleteIf(cond func(key K, v V) bool) {
	s.Range(func(key K, v V) bool {
		if cond(key, v) {
			s.Delete(key)
		}
		return true
	})
}
