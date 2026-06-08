package objects

import (
	"maps"
	"sync"
)

/*
	REDO to recycle removed ids allowing to stay in uint16 ids = less bandwith, server can run forever,
	atm uint64 is alot + after working for a while we can reach id cap
*/

type SharedCollection[T any] struct {
	objectsMap map[uint64]T
	nextID     uint64
	mapMux     sync.Mutex
}

func NewSharedCollection[T any](capacity ...int) *SharedCollection[T] {
	var newObjMap map[uint64]T

	if len(capacity) > 0 {
		newObjMap = make(map[uint64]T, capacity[0])
	} else {
		newObjMap = make(map[uint64]T)
	}

	return &SharedCollection[T]{
		objectsMap: newObjMap,
		nextID:     1,
	}
}

func (s *SharedCollection[T]) Add(obj T, id ...uint64) uint64 {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()

	thisId := s.nextID
	if len(id) > 0 {
		thisId = id[0]
	}

	s.objectsMap[thisId] = obj
	s.nextID++
	return thisId
}

func (s *SharedCollection[T]) Remove(id uint64) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()
	delete(s.objectsMap, id)
}

func (s *SharedCollection[T]) ForEach(callback func(uint64, T)) {
	s.mapMux.Lock()
	copy := make(map[uint64]T, len(s.objectsMap))
	maps.Copy(copy, s.objectsMap)
	s.mapMux.Unlock()

	for id, obj := range copy {
		callback(id, obj)
	}
}

func (s *SharedCollection[T]) Get(id uint64) (T, bool) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()
	obj, ok := s.objectsMap[id]
	return obj, ok
}

func (s *SharedCollection[T]) Len() int {
	return len(s.objectsMap)
}