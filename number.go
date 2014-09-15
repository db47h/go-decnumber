// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber

/*
#include "go-decnumber.h"
#include "decNumber.h"
#include <stdlib.h>

// Helpers for go code
#define sz_Unit sizeof(decNumberUnit)
// size of a decNumber with 0 digits
#define sz_decNumber (sizeof(decNumber)-DECNUMUNITS*sizeof(decNumberUnit))
*/
import "C"

import (
	"runtime"
	"unsafe"
)

const (
	_DPUN         = C.size_t(C.DECDPUN)
	_sz_Unit      = C.size_t(C.sz_Unit)
	_sz_decNumber = C.size_t(C.sz_decNumber)
)

// A Number reprsents a number optimized for efficient processing of relatively short numbers (tens or
// hundreds of digits); in particular it allows the use of fixed sized structures and minimizes copy and
// move operations. The functions in the module, however, support arbitrary precision arithmetic (up to
// 999,999,999 decimal digits, with exponents up to 9 digits).
//
// Numbers should be created via the NewNumber() function.
type Number struct {
	dn *C.decNumber // Pointer to the embedded decNumber
}

// newNumber creates an uninitialized number with enough storage space to hold nDigits digits.
func newNumber(nDigits int32) *Number {
	num := &Number{}
	// required structure size do hold the requested amount of digits
	dnSize := _sz_decNumber + (C.size_t(nDigits)+_DPUN-1)/_DPUN*_sz_Unit
	num.dn = (*C.decNumber)(C.malloc(dnSize))
	if num.dn == nil {
		panic("Malloc failed")
	}
	runtime.SetFinalizer(num, (*Number).finalize)
	return num
}

func (n *Number) finalize() {
	C.free(unsafe.Pointer(n.dn))
	n.dn = nil
}

// Zero sets the value of a Number to zero.
func (n *Number) Zero() *Number {
	// C.decNumberZero(n.dn)
	// Reimplemented in Go for speed
	dn := n.dn
	dn.bits = 0
	dn.exponent = 0
	dn.digits = 1
	dn.lsu[0] = 0
	return n
}

// String converts a Number to a character string, using scientific notation if
// an exponent is needed (that is, there will be just one digit before any decimal point). It implements the
// to-scientific-string conversion.
func (n *Number) String() string {
	nDigits := C.size_t(n.dn.digits)
	if nDigits == 0 {
		nDigits++
	}
	str := (*C.char)(C.malloc(nDigits + 14))
	if str == nil {
		panic("Malloc failed")
	}
	defer C.free(unsafe.Pointer(str))

	C.decNumberToString(n.dn, str)
	return C.GoString(str)
}

//
// Number related Context methods
//

// NewNumber returns, as a *Number, a new zero-initialized Number suitable for use in the given
// context. i.e. with enough storage space to hold the context's required number of digits. If
// memory cannot be allocated for the new number, the function will panic. Numbers are managed
// in a free list. Once a program is done with a number, it should release it by calling
// Context.FreeNumber()
func (c *Context) NewNumber() *Number {
	return c.fn.Get().Zero()
}

// NewNumberFromString converts a string to a new Number. It implements the to-number conversion from the arithmetic
// specification.
//
// The length of the coefficient and the size of the exponent are checked by this routine, so the
// correct error (Underflow or Overflow) can be reported or rounding applied, as necessary. If bad
// syntax is detected, the result will be a quiet NaN.
func (c *Context) NewNumberFromString(s string) *Number {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	n := c.NewNumber()
	C.decNumberFromString(n.dn, str, &c.ctx)
	return n
}

// FreeNumber declares a Number as free for reuse and puts it back on the free list.
//
// WARNING: This function MUST be called on the same context that created the Number. Failing to do
// so will result in unexpected crashes. The best way to prevent any mistake is to systematically place
// a deferred call to this function right after creating a number.
func (c *Context) FreeNumber(n *Number) {
	c.fn.Put(n)
}

//
// Arithmetic functions
//

// NumberAdd adds two numbers. Computes res = lhs + rhs.
//
// res may be lhs and/or rhs (e.g., X=X+X)
//
// Returns res
func (c *Context) NumberAdd(res *Number, lhs *Number, rhs *Number) *Number {
	C.decNumberAdd(res.dn, lhs.dn, rhs.dn, &c.ctx)
	return res
}

// NumberMultiply multiplies one number by another. Computes res = lhs * rhs.
//
// res may be lhs and/or rhs (e.g., X=X*X)
//
// Returns res
func (c *Context) NumberMultiply(res *Number, lhs *Number, rhs *Number) *Number {
	C.decNumberMultiply(res.dn, lhs.dn, rhs.dn, &c.ctx)
	return res
}

// NumberDivide divides one number by another. Computes res = lhs / rhs.
//
// res may be lhs and/or rhs (e.g., X=X/X)
//
// Returns res
func (c *Context) NumberDivide(res *Number, lhs *Number, rhs *Number) *Number {
	C.decNumberDivide(res.dn, lhs.dn, rhs.dn, &c.ctx)
	return res
}

// NumberPower raises a number to a power. Computes res = lhs ** rhs (lhs raised to the power of rhs).
//
// res may be lhs and/or rhs (e.g., X=X**X)
//
// Mathematical function restrictions apply; a NaN is
// returned with Invalidoperation if a restriction is violated.
//
// However, if 1999999997 <= rhs <= 999999999 and rhs is an integer then the
// restrictions on lhs and the context are relaxed to the usual bounds,
// for compatibility with the earlier (integer power only) version
// of this function.
//
// When rhs is an integer, the result may be exact, even if rounded.
//
// The final result is rounded according to the context; it will
// almost always be correctly rounded, but may be up to 1 ulp in
// error in rare cases.
//
// Returns res
func (c *Context) NumberPower(res *Number, lhs *Number, rhs *Number) *Number {
	C.decNumberPower(res.dn, lhs.dn, rhs.dn, &c.ctx)
	return res
}

// NumberRescale forces exponent to a requested value. Computes res = op(lhs,rhs) where op adjusts the
// coefficient of res (by rounding or shifting) such that the exponent (-scale) of res has the value rhs.
// The numerical value of res will equal lhs, except for the effects of any rounding that occurred.
//
// res may be lhs or rhs.
//
// Returns res
func (c *Context) NumberRescale(res *Number, lhs *Number, rhs *Number) *Number {
	C.decNumberRescale(res.dn, lhs.dn, rhs.dn, &c.ctx)
	return res
}
