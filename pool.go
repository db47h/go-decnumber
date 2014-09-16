// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber

// poolSize is the default maximum size for new pools.
var poolSize uint = 128

// A Pool is a set of temporary objects that may be individually saved and retrieved.
//
// Pool's purpose is to cache allocated but unused items for later reuse, relieving pressure on the
// garbage collector.
//
// This is a naïve implementation based on a fixed capacity Go channel. It is not thread-safe and is
// only provided as a lightweight alternative to sync.Pool to manage free-lists of *Number's. See
// NumberPool().
type Pool struct {
	pool    chan interface{}   // use a channel to hold pooled data
	New     func() interface{} // how to create new items
	MaxSize uint               // maximum size of the pool
}

// initPool initializes the pool on first use
func (p *Pool) initPool() {
	size := p.MaxSize
	if size == 0 {
		size = poolSize
	}
	p.pool = make(chan interface{}, size)
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
	select {
	case item := <-p.pool:
		return item
	default:
	}
	if p.New != nil {
		return p.New()
	}
	return nil
}

// Put adds x to the pool. If the pool is full, x will just get discarded silently.
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}
	if p.pool == nil {
		p.initPool()
	}
	select {
	case p.pool <- x:
	default:
	}
}
