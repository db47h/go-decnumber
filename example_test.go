// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	"."
	"fmt"
	"sync"
)

// Go re-implementation of decNumber's example1.c - simple addition.
//
// Cnvert the first two argument words to decNumber, add them together, and display the result.
func Example_example1() {
	var (
		arg1         = "1.27"
		arg2         = "2.23"
		digits int32 = 34
	)
	ctx := dec.NewContext(dec.InitBase, digits)

	a := dec.NewNumber(digits).FromString(arg1, ctx) // Should not ignore errors...
	b := dec.NewNumber(digits).FromString(arg2, ctx)

	a.Add(a, b, ctx) // a=a+b

	// there is no need to call c.String(). fmt.Print("%s") takes care of it.
	fmt.Printf("%s + %s => %s\n", arg1, arg2, a)

	// Output:
	// 1.27 + 2.23 => 3.50
}

// NewNumber() example
func Example_NewNumber() {
	// create a context with 99 digits precision, just for kicks
	ctx := dec.NewContext(dec.InitBase, 99)
	// create a number
	n := dec.NewNumber(ctx.Digits())

	// an IEEE 754 decimal128 type context
	// using the default 34 digits precision
	ctx = dec.NewContext(dec.InitDecimal128, 0)
	n = dec.NewNumber(ctx.Digits())
	// Set it to zero
	n.Zero()
	fmt.Println(n)

	// Output:
	// 0
}

// Accpeted formats and error handling demo.
func ExampleNumber_FromString() {
	// new context
	ctx := dec.NewContext(dec.InitDecimal64, 0)
	// We're lazy, and since we can do it, define a shorthand
	New := func(s string) *dec.Number {
		return dec.NewNumber(ctx.Digits()).FromString(s, ctx)
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
	ctx := dec.NewContext(dec.InitDecimal128, 0)

	// New() function for the pool to create new numbers
	newFunc := func() interface{} { return dec.NewNumber(ctx.Digits()) }

	// create a pool. Either dec.Pool or sync.Pool will do
	syncPool := sync.Pool{New: newFunc}

	// We can use Get().(*dec.Number) to get new or reusable numbers
	number := syncPool.Get().(*dec.Number)
	fmt.Printf("from sync.Pool: %s\n", number.Zero())
	// We're done with it, put it back in the pool
	syncPool.Put(number)

	// Or, wrap it with a NumberPool so that Get() returns *Number instead of interface{}.
	// NumberPool also helps keeping track of the context.
	pool := &dec.NumberPool{&syncPool, ctx}
	// and benefit: no need to type-cast
	number = pool.Get()
	// Introducing the idiomatic code: defer Put() the *Number right after Get()
	defer pool.Put(number)
	fmt.Printf("from sync.Pool: %s\n", number.FromString("1243", pool.Context))

	// Output:
	// from sync.Pool: 0
	// from sync.Pool: 1243
}

// Compact version of example 1, using dec.Pool
func ExampleNumberPool_2() {
	// Create a Context
	ctx := dec.NewContext(dec.InitDecimal128, 0)

	// And a usable pool based on dec.Pool
	pool := &dec.NumberPool{
		&dec.Pool{
			New: func() interface{} { return dec.NewNumber(ctx.Digits()) },
		},
		ctx,
	}

	// Now create numbers
	number := pool.Get()
	defer pool.Put(number)
	fmt.Printf("from dec.Pool: %s\n", number.Zero())

	number = pool.Get()
	defer pool.Put(number)
	fmt.Printf("from dec.Pool: %s\n", number.FromString("1243", pool.Context))

	// Output:
	// from dec.Pool: 0
	// from dec.Pool: 1243
}
