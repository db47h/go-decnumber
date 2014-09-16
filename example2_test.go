// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "." // "github.com/wildservices/go-decnumber"
	"fmt"
	"sync"
)

var (
	// Global context
	gCtx = dec.NewContext(dec.InitBase, 25) // 25 digits

	// Some constants
	one     = dec.NewNumber(gCtx).FromString("1", gCtx)   // 1
	mTwo    = dec.NewNumber(gCtx).FromString("-2", gCtx)  // -2
	hundred = dec.NewNumber(gCtx).FromString("100", gCtx) // 100
)

// CompoundInterest calculates compound interests.
//
// Arguments are *NumberPool, investment, rate (%), and years
func CompoundInterest(p *dec.NumberPool, start *dec.Number, rate *dec.Number, years *dec.Number) (*dec.Number, error) {
	// Assume that we don't have a global context, so we use the *NumberPool
	// to pass around a valid Context
	ctx := p.Context
	// save status and clear
	// leverage promotion of Context methods to NumberPool methods
	saved := p.Status().Save(dec.Errors)
	p.Status().Clear(dec.Errors)

	// Compute
	t := p.Get()                             // get a temporary number t
	defer p.Put(t)                           // put back t when we're done with it
	t.Divide(rate, hundred, ctx)             // t=rate/100
	t.Add(t, one, ctx)                       // t=t+1
	t.Power(t, years, ctx)                   // t=t**years
	total := p.Get().Multiply(t, start, ctx) // total=t*start (total created on the fly)
	total.Rescale(total, mTwo, ctx)          // two digits please

	// Function epilogue:
	err := p.ErrorStatus() // check errors
	// Merge previous status
	p.Status().Set(*saved)

	return total, err

}

// Extended re-implementation of decNumber's example2.c.
func Example_example2() {

	// This is our main function where we setup a NumberPool and collect
	// arguments.

	// Create a global NumberPool
	p := &dec.NumberPool{
		&sync.Pool{New: func() interface{} { return dec.NewNumber(gCtx) }},
		gCtx,
	}
	// arguments for CompoundInterest
	var (
		start = p.Get().FromString("50000", p.Context)
		rate  = p.Get().FromString("3.17", p.Context)
		years = p.Get().FromString("12", p.Context)
	)
	// dispose of them when done.
	defer func() {
		p.Put(start)
		p.Put(rate)
		p.Put(years)
	}()

	// do some computing at last
	total, err := CompoundInterest(p, start, rate, years)

	// error check ?
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// print results
	fmt.Printf("%s at %s%% for %s years => %s\n", start, rate, years, total)
	// and dispose of the result
	defer p.Put(total)

	// Output:
	// 50000 at 3.17% for 12 years => 72712.85
}
