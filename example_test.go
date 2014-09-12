// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	"bitbucket.org/wildservices/go-decnumber"
	"fmt"
)

// NewNumber() example
func ExampleNewNumber() {
	// create a context with 99 digits precision
	ctx := decnumber.NewCustomContext(99, decnumber.MaxEMax, decnumber.MinEMin, decnumber.RoundHalfEven, 0)

	// create a number
	n := ctx.NewNumber()
	defer ctx.FreeNumber(n) // idiomatic deferred call to FreeNumber()

	// an IEEE 754 decimal128 type context
	ctx = decnumber.NewContext(decnumber.InitDecimal128)
	n = ctx.NewNumber()
	defer ctx.FreeNumber(n)
}

// Various examples of Number.NumberFromString()
func ExampleContext_NumberFromString() {
	ctx := decnumber.NewContext(decnumber.InitDecimal64)
	n, err := ctx.NumberFromString("378.2654651646516165416165315131232")
	defer ctx.FreeNumber(n)
	if err != nil {
		fmt.Println(err.Error)
	}
	// Decimal64 has only 16 digits, the number will be truncated to the context's precision
	fmt.Printf("%s\n", n.String())

	// infinite number
	// Since NumberFromString may change the Context status, we chain call
	// NumberFromString() with ZeroStatus()
	n, err = ctx.ZeroStatus().NumberFromString("-INF")
	defer ctx.FreeNumber(n)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Scientific notation
	n, err = ctx.ZeroStatus().NumberFromString("1.275654e16")
	defer ctx.FreeNumber(n)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// error. Will cause an overflow and set the number to +Infinity
	// This is still a "valid" number for some applications
	n, err = ctx.ZeroStatus().NumberFromString("1.275654e321455")
	defer ctx.FreeNumber(n)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s\n", n.String())

	// Here, we will get a conversion syntax error
	// and the number will be set to NaN (not a number)
	n, err = ctx.ZeroStatus().NumberFromString("12garbage524")
	defer ctx.FreeNumber(n)
	if err != nil {
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
