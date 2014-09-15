// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	"."
	"fmt"
)

// Go re-implementation of decNumber's example1.c.
//
// Cnvert the first two argument words to decNumber, add them together, and display the result
func Example_example1() {
	arg1 := "1.27"
	arg2 := "2.23"

	ctx := decnumber.NewContext(decnumber.InitBase, 34)

	a := ctx.NewNumber().FromString(arg1) // Should not ignore errors...
	b := ctx.NewNumber().FromString(arg2)

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
	// func CompoundInterest(ctx *Context, start *Number, rate *Number, years *Number) (*Number, error)
	//
	// Simulate arguments passed to the function
	var (
		ctx   = decnumber.NewContext(decnumber.InitBase, 25)
		start = ctx.NewNumber().FromString("50000")
		rate  = ctx.NewNumber().FromString("3.17")
		years = ctx.NewNumber().FromString("12")
	)

	// Real function start:
	var (
		one     = ctx.NewNumber().FromString("1")
		mTwo    = ctx.NewNumber().FromString("-2")
		hundred = ctx.NewNumber().FromString("100")
	)
	defer func() {
		one.Release()
		mTwo.Release()
		hundred.Release()
	}()
	// The above constants should really be taken from a const library

	// save status and clear
	saved := ctx.Status().Save(decnumber.Errors)
	ctx.Status().Clear(decnumber.Errors)

	// Compute
	t := ctx.NewNumber()                        // get a temporary number t
	defer t.Release()                           // Release(t) when we're done
	t.Divide(rate, hundred)                     // t=rate/100
	t.Add(t, one)                               // t=t+1
	t.Power(t, years)                           // t=t**years
	total := ctx.NewNumber().Multiply(t, start) // total=t*start
	total.Rescale(total, mTwo)                  // two digits please
	// do not Release(total) since we return it as a result. The caller may however do so.

	// Function epilogue:
	_ = /* err := */ ctx.ErrorStatus() // check errors
	// Merge previous status
	ctx.Status().Set(*saved)

	fmt.Printf("%s at %s%% for %s years => %s\n",
		start, rate, years, total)

	// return total, err

	// Output:
	// 50000 at 3.17% for 12 years => 72712.85
}

// NewNumber() example
func ExampleContext_NewNumber() {
	// create a context with 99 digits precision, just for kicks
	ctx := decnumber.NewContext(decnumber.InitBase, 99)

	// create a number
	n := ctx.NewNumber()
	defer n.Release() // idiomatic deferred call to Release()

	// an IEEE 754 decimal128 type context
	ctx = decnumber.NewContext(decnumber.InitDecimal128, 0)
	n = ctx.NewNumber()
	defer n.Release()
}

// Accpeted formats and error handling demo.
func ExampleNumber_FromString() {
	ctx := decnumber.NewContext(decnumber.InitDecimal64, 0)
	n := ctx.NewNumber().FromString("378.2654651646516165416165315131232")
	defer n.Release()
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error)
	}
	// Decimal64 has only 16 digits, the number will be truncated to the context's precision
	fmt.Printf("%s\n", n.String())

	// infinite number
	n = ctx.NewNumber().FromString("-INF")
	defer n.Release()
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Scientific notation
	n = ctx.NewNumber().FromString("1.275654e16")
	defer n.Release()
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// error. Will cause an overflow and set the number to +Infinity
	// This is still a "valid" number for some applications
	n = ctx.NewNumber().FromString("1.275654e321455")
	defer n.Release()
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Here, we will get a conversion syntax error
	// and the number will be set to NaN (not a number)
	//
	// We call ZeroStatus() to clear any previous error
	n = ctx.ZeroStatus().NewNumber().FromString("12garbage524")
	defer n.Release()
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
