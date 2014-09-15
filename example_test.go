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

	ctx := decnumber.NewContext(decnumber.InitBase)

	a := ctx.NewNumberFromString(arg1) // Should not ignore errors...
	b := ctx.NewNumberFromString(arg2)

	// Not in the original example: error checking.
	// If an error occured while converting either a or b, err will be set to a non nil value
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	ctx.NumberAdd(a, a, b) // a=a+b

	// there is no need to call c.String(). fmt.Print("%s") takes care of it.
	fmt.Printf("%s + %s => %s\n", arg1, arg2, a)

	// Output:
	// 1.27 + 2.23 => 3.50
}

// Go re-implementation of decNumber's example1.c.
//
// Calculate compound interest.
// Arguments are investment, rate (%), and years
func Example_example2() {
	arg1 := "50000"
	arg2 := "3.17"
	arg3 := "12"

	ctx := decnumber.NewCustomContext(25, decnumber.MaxMath, -decnumber.MaxMath,
		decnumber.RoundHalfEven, 0)

	one := ctx.NewNumberFromString("1")
	mTwo := ctx.NewNumberFromString("-2")
	hundred := ctx.NewNumberFromString("100")

	start := ctx.NewNumberFromString(arg1)
	rate := ctx.NewNumberFromString(arg2)
	years := ctx.NewNumberFromString(arg3)

	// Not in the original example: error checking.
	// If an error occured while converting the arguments, err will be set to a non nil value
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	ctx.NumberDivide(rate, rate, hundred)                     // rate=rate/100
	ctx.NumberAdd(rate, rate, one)                            // rate=rate+1
	ctx.NumberPower(rate, rate, years)                        // rate=rate**years
	total := ctx.NumberMultiply(ctx.NewNumber(), rate, start) // total=rate*start
	total = ctx.NumberRescale(total, total, mTwo)             // two digits please

	fmt.Printf("%s at %s%% for %s years => %s\n",
		arg1, arg2, arg3, total)

	// Output:
	// 50000 at 3.17% for 12 years => 72712.85
}

// NewNumber() example
func ExampleContext_NewNumber() {
	// create a context with 99 digits precision
	ctx := decnumber.NewCustomContext(99, decnumber.MaxMath, 1-decnumber.MaxMath, decnumber.RoundHalfEven, 0)

	// create a number
	n := ctx.NewNumber()
	defer ctx.FreeNumber(n) // idiomatic deferred call to FreeNumber()

	// an IEEE 754 decimal128 type context
	ctx = decnumber.NewContext(decnumber.InitDecimal128)
	n = ctx.NewNumber()
	defer ctx.FreeNumber(n)
}

// Accpeted formats and error handling demo.
func ExampleContext_NewNumberFromString() {
	ctx := decnumber.NewContext(decnumber.InitDecimal64)
	n := ctx.NewNumberFromString("378.2654651646516165416165315131232")
	defer ctx.FreeNumber(n)
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error)
	}
	// Decimal64 has only 16 digits, the number will be truncated to the context's precision
	fmt.Printf("%s\n", n.String())

	// infinite number
	n = ctx.NewNumberFromString("-INF")
	defer ctx.FreeNumber(n)
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Scientific notation
	n = ctx.NewNumberFromString("1.275654e16")
	defer ctx.FreeNumber(n)
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// error. Will cause an overflow and set the number to +Infinity
	// This is still a "valid" number for some applications
	n = ctx.NewNumberFromString("1.275654e321455")
	defer ctx.FreeNumber(n)
	if err := ctx.ErrorStatus(); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Here, we will get a conversion syntax error
	// and the number will be set to NaN (not a number)
	//
	// We call ZeroStatus() to clear any previous error
	n = ctx.ZeroStatus().NewNumberFromString("12garbage524")
	defer ctx.FreeNumber(n)
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
