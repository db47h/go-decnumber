// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	"."
	"./util"
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

// Go re-implementation of decNumber's example5.c. Compressed formats.
func Example_example5() {
	var (
		a     = new(dec.Quad) // Should be dec.Double, but not implemented yet
		d     *dec.Number
		ctx   *dec.Context
		hexes string
	)

	ctx = dec.NewContext(dec.InitDecimal128, 0) // will be 16 digits

	a.FromString("127.9984", ctx)
	// lay out the Quad as hexadecimal pairs
	// big endian - ordered
	for _, b := range a.Bytes() {
		if dec.LittleEndian {
			hexes = fmt.Sprintf("%02X ", b) + hexes
		} else {
			hexes += fmt.Sprintf("%02X ", b)
		}
	}

	d = a.ToNumber(nil)
	fmt.Printf("%s => %s=> %s\n", a, hexes, d)

	// Output:
	// 127.9984 => 22 07 00 00 00 00 00 00 00 00 00 00 00 15 E6 8E => 127.9984
}

// Go re-implementation of decNumber's example6.c. Packed Decimal numbers.
//
// This example reworks Example 2, starting and ending with Packed Decimal numbers.
func Example_example6() {

	// This is our main function where we setup a NumberPool and collect
	// arguments.

	// Create a global NumberPool
	p := &dec.NumberPool{
		&util.Pool{New: func() interface{} { return dec.NewNumber(gCtx.Digits()) }},
		gCtx,
	}
	// arguments for CompoundInterest
	var (
		startp   = dec.Packed{[]byte{0x5C}, -4}      // 5e+4 = 50000
		ratep    = dec.Packed{[]byte{0x31, 0x7C}, 2} // 3.17
		yearsp   = dec.Packed{[]byte{0x01, 0x2C}, 0} // 12
		start, _ = startp.ToNumber(nil)
		rate, _  = ratep.ToNumber(nil)
		years, _ = yearsp.ToNumber(nil)
	)

	// do some computing at last
	total, err := CompoundInterest(p, start, rate, years)

	// error check ?
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// print results
	fmt.Printf("%s at %s%% for %s years => %s\n", start, rate, years, total)
	resPacked := new(dec.Packed)
	resPacked.FromNumber(total)
	// and dispose of the result
	defer p.Put(total)

	for _, d := range resPacked.Buf {
		fmt.Printf("%02X ", d)
	}
	fmt.Printf("(scale: %d)\n", resPacked.Scale)

	// Output:
	// 5E+4 at 3.17% for 12 years => 72712.85
	// 72 71 28 5C (scale: 2)
}

// Go re-implementation of decNumber's example7.c. Using decQuad to add two numbers together.
func Example_example7() {
	var (
		a   = new(dec.Quad)
		b   = new(dec.Quad)
		ctx *dec.Context
	)

	// Context suitable for Quads
	ctx = dec.NewContext(dec.InitQuad, 0)

	a.FromString("123.456", ctx)
	b.FromString("7890.12", ctx)
	a.Add(a, b, ctx) // a = a + b

	s := a.String()

	fmt.Printf("123.456 + %s => %s\n", b, s)

	// Output:
	// 123.456 + 7890.12 => 8013.576
}

// Go re-implementation of decNumber's example8.c.  Using Quad with Number
func Example_example8() {
	var (
		a          = new(dec.Quad)
		numa, numb *dec.Number
		ctx        = dec.NewContext(dec.InitQuad, 0) // Initialize
	)

	a.FromString("1234.567", ctx) // get a
	as := a.String()              // keep a string copy for test output
	a.Add(a, a, ctx)              // double a
	numa = a.ToNumber(nil)        // convert to Number
	numb = dec.NewNumber(ctx.Digits()).FromString("98.7654", ctx)
	numa.Power(numa, numb, ctx) // numa=numa**numb
	a.FromNumber(numa, ctx)     // back to quad

	fmt.Printf("power(2*%s, %s) => %s \n", as, numb, a)

	// Output:
	// power(2*1234.567, 98.7654) => 1.164207353496260978533862797373143E+335
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
		&util.Pool{
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
