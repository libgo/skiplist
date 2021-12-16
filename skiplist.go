// Copyright (c) 2021, Rod Dong <rod.dong@me.com> All rights reserved.
//
// Use of this source code is governed by The MIT License.

// skiplist implments a redis-like skiplist structure in memory.
package skiplist

import (
	"constraints"
	"errors"
	"math/rand"
	"time"
)

const (
	MaxLevel = 12
	P        = 0.3
)

type skiplistNode[K constraints.Ordered, V any] struct {
	key      K
	value    V
	level    int
	forward  []*skiplistNode[K, V]
	backward *skiplistNode[K, V]
}

type Skiplist[K constraints.Ordered, V any] struct {
	header *skiplistNode[K, V]
	tail   *skiplistNode[K, V]
	level  int
	length int
}

// New a empty skiplist, the zeroK & zeroV is used for nil/default value.
// Not familiar with generic, zeroK & zeroV should be modified later.
func New[K constraints.Ordered, V any]() *Skiplist[K, V] {
	var zeroK K
	var zeroV V
	header := &skiplistNode[K, V]{
		zeroK,
		zeroV,
		MaxLevel,
		make([]*skiplistNode[K, V], MaxLevel),
		nil,
	}
	return &Skiplist[K, V]{
		header,
		nil,
		1,
		0,
	}
}

// Put a new key/value into skiplist. If exists, update the value.
func (sl *Skiplist[K, V]) Put(k K, v V) error {
	// if this is the first element, just insert into level 0
	if sl.length == 0 {
		n := &skiplistNode[K, V]{
			key:      k,
			value:    v,
			level:    1,
			forward:  make([]*skiplistNode[K, V], 1),
			backward: sl.header,
		}
		sl.header.forward[0] = n
		sl.tail = n
		sl.length = 1
		return nil
	}

	// find the position to update/insert key+value
	found, node, updates := sl.find(k)

	// if found, update value
	if found {
		node.value = v
		return nil
	}

	// if not found, insert a new node
	level := randomLevel()
	n := &skiplistNode[K, V]{
		key:      k,
		value:    v,
		level:    level,
		forward:  make([]*skiplistNode[K, V], level),
		backward: nil,
	}

	for i := 0; i < level; i++ {
		n.forward[i] = updates[i].forward[i]
		updates[i].forward[i] = n
	}

	n.backward = updates[0]
	if n.forward[0] != nil {
		n.forward[0].backward = n
	} else {
		sl.tail = n
	}

	sl.length++
	return nil
}

// Get return the value of key. If not found, error(Not Found).
func (sl *Skiplist[K, V]) Get(key K) (V, error) {
	found, node, _ := sl.find(key)
	if found {
		return node.value, nil
	}
	var zeroV V
	return zeroV, errors.New("Not Found")
}

// Del remove the key from skiplist.
func (sl *Skiplist[K, V]) Del(key K) error {
	found, node, updates := sl.find(key)
	if !found {
		return errors.New("Not Found")
	}

	if node.forward[0] != nil {
		node.forward[0].backward = node.backward
	} else {
		sl.tail = node.backward
	}

	for i := 0; i < node.level; i++ {
		updates[i].forward[i] = node.forward[i]
	}
	sl.length--
	return nil
}

// Length return the length of skiplist.
func (sl *Skiplist[K, V]) Length() int {
	return sl.length
}

// RangeByKey return range query with start key and end key.
func (sl *Skiplist[K, V]) RangeByKey(start K, end K) (map[K]V, error) {
	if start > end {
		return nil, errors.New("START key is great than END key")
	}

	result := make(map[K]V)
	found, node, updates := sl.find(start)
	if !found {
		node = updates[0].forward[0]
	}

	for ; node != nil && node.key <= end; node = node.forward[0] {
		result[node.key] = node.value
	}

	return result, nil
}

// RangeByCount return range query with start and count.
func (sl *Skiplist[K, V]) RangeByCount(start K, count int) (map[K]V, error) {
	if count == 0 {
		return nil, errors.New("Zero COUNT")
	}

	// count < 0 means backward query
	forward := true
	if count < 0 {
		count = 0 - count
		forward = false
	}

	result := make(map[K]V)
	found, node, updates := sl.find(start)

	// If not found, set the node to updates[0].forward[0] when count>=0, or to updates[0] when count<0
	if !found {
		if forward {
			node = updates[0].forward[0]
		} else {
			node = updates[0]
		}
	}

	// Get the query result
	for c := 0; node != nil && node != sl.header && c < count; c++ {
		result[node.key] = node.value
		if forward {
			node = node.forward[0]
		} else {
			node = node.backward
		}
	}

	return result, nil
}

// RangeByIndex return range query with start and count.
func (sl *Skiplist[K, V]) RangeByIndex(start int, count int) (map[K]V, error) {
	if count == 0 {
		return nil, errors.New("Zero COUNT")
	}
	if start >= sl.length || start < 0-sl.length {
		return nil, errors.New("Out of range")
	}

	// count < 0 means backward query
	forward := true
	if count < 0 {
		count = 0 - count
		forward = false
	}

	result := make(map[K]V)

	// Find the START node
	node := sl.header.forward[0]
	if start >= 0 {
		for i := 0; node != nil && i < start; i++ {
			node = node.forward[0]
		}
	} else {
		node = sl.tail
		for i := -1; node != sl.header && i > start; i-- {
			node = node.backward
		}
	}

	// Get the query result
	for c := 0; node != nil && node != sl.header && c < count; c++ {
		result[node.key] = node.value
		if forward {
			node = node.forward[0]
		} else {
			node = node.backward
		}
	}

	return result, nil
}

// find the key from Skiplist, and try to return update nodes to insert/delete.
func (sl *Skiplist[K, V]) find(key K) (found bool, node *skiplistNode[K, V], updates []*skiplistNode[K, V]) {
	updates = make([]*skiplistNode[K, V], MaxLevel)

	c := sl.header
	found = false
	node = nil

	for i := MaxLevel - 1; i >= 0; i-- {
		for ; c.forward[i] != nil && c.forward[i].key < key; c = c.forward[i] {
			// Forward to the node
		}

		// Found
		if c.forward[i] != nil && c.forward[i].key == key {
			found = true
			node = c.forward[i]
		}
		updates[i] = c
	}
	return
}

// randomLevel generates skiplist node level.
func randomLevel() int {
	level := 1
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		v := r.Uint32()
		if float32(v&0xFFFF) > float32(P*0xFFFF) {
			break
		}
		level++
	}
	if level > MaxLevel {
		return MaxLevel
	}
	return level
}
