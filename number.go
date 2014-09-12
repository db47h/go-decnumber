// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* This file is a wrapper around decNumber.c */

package decnumber

/*
#cgo CFLAGS: -Ilibdecnumber

#include "go-decnumber.h"
#include "decNumber.h"
#include <stdlib.h>

#include "decNumber.c"

size_t decNumberStructSize(int32_t digits) {
	return sizeof(decNumber)+(D2U(digits-1)*sizeof(Unit));
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// A Number reprsents a number optimized for efficient processing of relatively short numbers (tens or
// hundreds of digits); in particular it allows the use of fixed sized structures and minimizes copy and
// move operations. The functions in the module, however, support arbitrary precision arithmetic (up to
// 999,999,999 decimal digits, with exponents up to 9 digits).
//
// Numbers should be created via the NewNumber() function.
type Number struct {
	n *C.decNumber // Pointer to the embedded decNumber
}

func newNumber(digits int32) *Number {
	// estimate necessary space to allocate structure
	num := &Number{}
	num.n = (*C.decNumber)(C.malloc(C.decNumberStructSize(C.int32_t(digits))))
	if num.n == nil {
		panic("Malloc failed")
	}
	runtime.SetFinalizer(num, (*Number).finalize)
	C.decNumberZero(num.n)
	return num
}

func (n *Number) finalize() {
	C.free(unsafe.Pointer(n.n))
	n.n = nil
}

// Zero sets the value of a Number to zero.
func (n *Number) Zero() *Number {
	C.decNumberZero(n.n)
	return n
}

// String converts a Number to a character string, using scientific notation if
// an exponent is needed (that is, there will be just one digit before any decimal point). It implements the
// to-scientific-string conversion.
func (n *Number) String() string {
	str := (*C.char)(C.malloc(C.DECNUMDIGITS + 14))
	if str == nil {
		panic("Malloc failed")
	}
	defer C.free(unsafe.Pointer(str))

	C.decNumberToString(n.n, str)
	return C.GoString(str)
}

//
// Number related Context methods
//

// NewNumber returns a new Number suitable for use in the given context. i.e. with enough storage space
// to hold the context's required number of digits. If memory cannot be allocated for the new number,
// the function will panic. Numbers are managed in a free list. Once a program is done with a number, it
// should release it by calling Context.FreeNumber()
func (c *Context) NewNumber() *Number {
	return c.fn.Get()
}

// NumberFromString converts a string to a new Number. It implements the to-number conversion from the arithmetic
// specification.
// The conversion is exact provided that the numeric string has no more significant digits than are
// specified in the Context and the adjusted exponent is in the range specified by the Context's EMin
// and EMax. If there are more digits in the string than specified in the Context, or the exponent is
// out of range, the value will be rounded as necessary using the Context rounding mode. The
// Context therefore determines the maximum precision for unrounded numbers.
func (c *Context) NumberFromString(s string) (*Number, error) {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	n := c.NewNumber()
	C.decNumberFromString(n.n, str, &c.ctx)
	return n, c.GetError()
}

// FreeNumber declares a Number as free for reuse and puts it back on the free list.
//
// WARNING: This function MUST be called on the same context that created the Number. Failing to do
// so will result in unexpected crashes. The best way to prevent any mistake is to systematically place
// deferred call to this function right after creating a number.
func (c *Context) FreeNumber(n *Number) {
	// zero it before pushing it back
	n.Zero()
	c.fn.Put(n)
}

// Math functions
