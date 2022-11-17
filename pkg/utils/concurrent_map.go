package utils

import (
	"fmt"
	"sync"
)

type ValueCreator func() (interface{}, error)

type ConcurrentMap struct {
	internalMap sync.Map
}

type onceWithValue struct {
	once  sync.Once
	value interface{}
	err   error
}

func (m *ConcurrentMap) GetOrCreate(key interface{}, valueCreator ValueCreator) (actual interface{}, err error) {
	
	res, _ := m.internalMap.LoadOrStore(key, &onceWithValue{})
	onceWithValue := res.(*onceWithValue)

	onceWithValue.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				onceWithValue.err = fmt.Errorf("recovered from panic in value creator: %#v", r)
				m.internalMap.Delete(key)
			}
		}()
		value, err := valueCreator()
		if err != nil {
			onceWithValue.err = err
			m.internalMap.Delete(key)
			return
		}

		onceWithValue.value = value
	})

	return onceWithValue.value, onceWithValue.err
}
