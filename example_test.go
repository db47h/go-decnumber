// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	"."
	"fmt"
	"sync"
)

// Go re-implementation of decNumber's example1.c.
//
// Cnvert the first two argument words to decNumber, add them together, and display the result
func Example_example1() {
	arg1 := "1.27"
	arg2 := "2.23"

	ctx := decnumber.NewContext(decnumber.InitBase, 34)

	a := decnumber.NewNumber(ctx).FromString(arg1) // Should not ignore errors...
	b := decnumber.NewNumber(ctx).FromString(arg2)

	// Not in the original example: error checking.
	// If an error occured while converting either a or b, err will be set to a non nil value
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	a.Add(a, b) // a=a+b

	// there is no need to call c.String(). fmt.Print("%s") takes care of it.
	fmt.Printf("%s + %s => %s\n", arg1, arg2, a)

	// Output:
	// 1.27 + 2.23 => 3.50
}

// Go re-implementation of decNumber's example2.c.
//
// Calculate compound interest.
// Arguments are investment, rate (%), and years
func Example_example2() {
	// While the original example was a kinda standalone demo, we turned this example
	// into what a real CompundInterest() function would look like. Real function prototype:
	//
	// func CompoundInterest(total *Number, start *Number, rate *Number, years *Number) error
	//
	// Simulate arguments passed to the function
	var (
		// The context and pool are usually global objects (or per goroutine)
		// A function-local pool would be almost pointless
		ctx   = decnumber.NewContext(decnumber.InitBase, 25)
		nlist = decnumber.NumberPool(&decnumber.Pool{
			New: func() interface{} { return decnumber.NewNumber(ctx) },
		})
		total = nlist.Get()
		start = nlist.Get().FromString("50000")
		rate  = nlist.Get().FromString("3.17")
		years = nlist.Get().FromString("12")
	)

	// Real function start:
	var (
		one     = nlist.Get().FromString("1")
		mTwo    = nlist.Get().FromString("-2")
		hundred = nlist.Get().FromString("100")
	)
	defer func() {
		nlist.Put(one)
		nlist.Put(mTwo)
		nlist.Put(hundred)
	}()
	// The above constants should really be taken from a const library

	// save status and clear
	saved := ctx.Status().Save(decnumber.Errors)
	ctx.Status().Clear(decnumber.Errors)

	// Compute
	t := nlist.Get()           // get a temporary number t
	defer nlist.Put(t)         // put back t when we're done with it
	t.Divide(rate, hundred)    // t=rate/100
	t.Add(t, one)              // t=t+1
	t.Power(t, years)          // t=t**years
	total.Multiply(t, start)   // total=t*start
	total.Rescale(total, mTwo) // two digits please

	// Function epilogue:
	_ = /* err := */ ctx.ErrorStatus() // check errors
	// Merge previous status
	ctx.Status().Set(*saved)

	fmt.Printf("%s at %s%% for %s years => %s\n",
		start, rate, years, total)

	// return err

	// Output:
	// 50000 at 3.17% for 12 years => 72712.85
}

// NewNumber() example
func Example_NewNumber() {
	// create a context with 99 digits precision, just for kicks
	ctx := decnumber.NewContext(decnumber.InitBase, 99)
	// create a number
	n := decnumber.NewNumber(ctx)

	// an IEEE 754 decimal128 type context
	// using the default 34 digits precision
	ctx = decnumber.NewContext(decnumber.InitDecimal128, 0)
	n = decnumber.NewNumber(ctx)
	// Set it to zero
	n.Zero()
	fmt.Println(n)

	// Output:
	// 0
}

// Accpeted formats and error handling demo.
func ExampleNumber_FromString() {
	// new context
	ctx := decnumber.NewContext(decnumber.InitDecimal64, 0)
	// We're lazy, and since we can do it, define a shorthand
	New := func(s string) *decnumber.Number {
		return decnumber.NewNumber(ctx).FromString(s)
	}

	n := New("378.2654651646516165416165315131232")
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error)
	}
	// Decimal64 has only 16 digits, the number will be truncated to the context's precision
	fmt.Printf("%s\n", n.String())

	// infinite number
	n = New("-INF")
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Scientific notation
	n = New("1.275654e16")
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// error. Will cause an overflow and set the number to +Infinity
	// This is still a "valid" number for some applications
	n = New("1.275654e321455")
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Here, we will get a conversion syntax error
	// and the number will be set to NaN (not a number)
	//
	// We call ZeroStatus() to clear any previous error
	ctx.ZeroStatus()
	n = New("12garbage524")
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Output:
	// 378.2654651646516
	// -Infinity
	// 1.275654E+16
	// Overflow
	// Infinity
	// Conversion syntax
	// NaN
}

// Example use of a pool to manage a free list of numbers
func ExampleNumberPool_1() {
	// Create a Context
	ctx := decnumber.NewContext(decnumber.InitDecimal128, 0)

	// New() function for the pool to create new numbers
	newFunc := func() interface{} { return decnumber.NewNumber(ctx) }

	// create a pool. Either decnumber.Pool or sync.Pool will do
	syncPool := sync.Pool{New: newFunc}

	// We can use Get().(*decnumber.Number) to get new or reusable numbers
	number := syncPool.Get().(*decnumber.Number)
	fmt.Printf("from sync.Pool: %s\n", number.Zero())
	// We're done with it, put it back in the pool
	syncPool.Put(number)

	// Or, wrap it with NumberPool() so that Get() returns *Number instead of interface{}
	pool := decnumber.NumberPool(&syncPool)
	// and benefit: no need to type-cast
	number = pool.Get()
	// Introducing the idiomatic code: defer Put() the *Number right after Get()
	defer pool.Put(number)
	fmt.Printf("from sync.Pool: %s\n", number.FromString("1243"))

	// Output:
	// from sync.Pool: 0
	// from sync.Pool: 1243
}

// Compact version of example 1, using decnumber.Pool
func ExampleNumberPool_2() {
	// Create a Context
	ctx := decnumber.NewContext(decnumber.InitDecimal128, 0)

	// And a usable pool based on decnumber.Pool
	pool := decnumber.NumberPool(&decnumber.Pool{
		New: func() interface{} { return decnumber.NewNumber(ctx) },
	})

	// Now create numbers
	number := pool.Get()
	defer pool.Put(number)
	fmt.Printf("from decnumber.Pool: %s\n", number.Zero())

	number = pool.Get()
	defer pool.Put(number)
	fmt.Printf("from decnumber.Pool: %s\n", number.FromString("1243"))

	// Output:
	// from decnumber.Pool: 0
	// from decnumber.Pool: 1243
}
