// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "." // "github.com/wildservices/go-decnumber"
	"fmt"
)

// Go re-implementation of decNumber's example6.c. Packed Decimal numbers.
//
// This example reworks Example 2, starting and ending with Packed Decimal numbers.
func Example_example6() {

	// This is our main function where we setup a NumberPool and collect
	// arguments.

	// Create a global NumberPool
	p := &dec.NumberPool{
		&dec.Pool{New: func() interface{} { return dec.NewNumber(gCtx.Digits()) }},
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
