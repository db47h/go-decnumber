// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package util provides a set of utility functions and types for the dec package.
//
package util

// poolSize is the default maximum size for new pools.
// Must be a power of two
var poolSize int = 128

// A Pool is a set of temporary objects that may be individually saved and retrieved.
//
// Pool's purpose is to cache allocated but unused items for later reuse, relieving pressure on the
// garbage collector.
//
// This is a naïve implementation based on a fixed capacity slice. It is not thread-safe and is
// only provided as a lightweight alternative to sync.Pool to manage free-lists of *Number's. See
// NumberPool().
type Pool struct {
	pool []interface{}      // use a channel to hold pooled data
	in   int                // Index of the next Put() value
	out  int                // Index of the next Get() value
	len  int                // number of items in the pool
	New  func() interface{} // how to create new items
}

// initPool initializes the pool on first use
func (p *Pool) initPool() {
	p.pool = make([]interface{}, poolSize)
}

// Get selects an arbitrary item from the Pool, removes it from the Pool, and returns it to the
// caller. Callers should not assume any relation between values passed to Put and the values
// returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns the result of calling p.New.
func (p *Pool) Get() interface{} {
	if p.pool == nil {
		p.initPool()
	}
	if p.len > 0 {
		v := p.pool[p.out]
		p.pool[p.out] = nil // remove reference
		p.out++
		p.out &= poolSize - 1 // => same as p.out %= poolSize
		p.len--
		return v
	}
	if p.New != nil {
		return p.New()
	}
	return nil
}

// Put adds x to the pool. If the pool is full, x will just get discarded silently.
func (p *Pool) Put(x interface{}) {
	if p.pool == nil {
		p.initPool()
	}
	if x == nil {
		return
	}
	if p.len == poolSize {
		return
	}
	p.pool[p.in] = x
	p.in++
	p.in &= poolSize - 1 // => same as p.in %= poolSize
	p.len++
}
