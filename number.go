// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#include "go-decnumber.h"
#include "decNumber.h"
#include <stdlib.h>

// Helpers for go code
decNumber * new_decNumber(int32_t digits) {
	return malloc( (sizeof(decNumber)-DECNUMUNITS*sizeof(decNumberUnit))
				+ ((size_t)digits+DECDPUN-1) / DECDPUN * sizeof(decNumberUnit) );
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
	dn *C.decNumber // Pointer to the embedded decNumber
}

// NewNumber returns, as a *Number, a new uinitialized Number with enough storage space for the
// requested number of digits. If memory cannot be allocated for the new number, the function will
// panic.
//
// Since the Number is unitialized, its value is not valid and must be initialized from some source
// before using it as an operand in an arithmetic operation. This is not necessary if the Number is
// to be used as the result of such operation.
func NewNumber(digits int32) *Number {
	num := &Number{}
	// required structure size do hold the requested amount of digits
	num.dn = C.new_decNumber(C.int32_t(digits))
	if num.dn == nil {
		panic("Malloc failed")
	}
	runtime.SetFinalizer(num, (*Number).finalize)
	return num
}

func (n *Number) finalize() {
	if n.dn != nil {
		C.free(unsafe.Pointer(n.dn))
		n.dn = nil
	}
}

// DecNumber returns a pointer to the underlying decNumber C struct.
func (n *Number) DecNumber() *C.decNumber {
	return n.dn
}

// Digits() returns the number of digits in a Number.
func (n *Number) Digits() int32 {
	return int32(n.dn.digits)
}

// Zero sets the value of a Number to zero.
func (n *Number) Zero() *Number {
	// C.decNumberZero(n.dn)
	// Reimplemented in Go for speed
	dn := n.dn
	dn.digits = 1
	dn.exponent = 0
	dn.bits = 0
	dn.lsu[0] = 0
	return n
}

// String converts a Number to a character string, using scientific notation if an exponent is
// needed (that is, there will be just one digit before any decimal point). It implements the
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

// FromString converts a string to a Number. It implements the to-number conversion from the
// arithmetic specification.
//
// The length of the coefficient and the size of the exponent are checked by this routine, so the
// correct error (Underflow or Overflow) can be reported or rounding applied, as necessary. If bad
// syntax is detected, the result will be a quiet NaN.
func (n *Number) FromString(s string, ctx *Context) *Number {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decNumberFromString(n.dn, str, ctx.DecContext())
	return n
}

//
// Pooling facilities
//

// A NumberPooler represents an object that can be used as a generic pool. sync.Pool and
// dec.Pool implement this interface.
type Pooler interface {
	Get() interface{}
	Put(interface{})
}

// A NumberPool wraps a Pooler to automatically type cast the result of Get() to a *Number.
// The *Context field is a convenience field to help in keeping track of the pool and associated
// Context with a single reference.
type NumberPool struct {
	Pooler
	*Context
}

// Get returns a free *Number from the pool.
func (p *NumberPool) Get() *Number {
	return p.Pooler.Get().(*Number)
}

// Put returns a *Number to the pool
// Not implemented. Uses promoted Pooler.Put()
// func (p *numberPool) Put(n *Number) {
//	p.Pooler.Put(n)
// }

//
// Arithmetic functions
//

// NumberAdd adds two numbers. Computes res = lhs + rhs.
//
// Returns res.
func (res *Number) Add(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberAdd(res.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return res
}

// NumberMultiply multiplies one number by another. Computes res = lhs * rhs.
//
// Returns res.
func (res *Number) Multiply(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberMultiply(res.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return res
}

// NumberDivide divides one number by another. Computes res = lhs / rhs.
//
// Returns res.
func (res *Number) Divide(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberDivide(res.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return res
}

// NumberPower raises a number to a power. Computes res = lhs ** rhs (lhs raised to the power of rhs).
//
// Mathematical function restrictions apply; a NaN is returned with Invalidoperation if a
// restriction is violated.
//
// However, if 1999999997 <= rhs <= 999999999 and rhs is an integer then the restrictions on lhs and
// the context are relaxed to the usual bounds, for compatibility with the earlier (integer power
// only) version of this function.
//
// When rhs is an integer, the result may be exact, even if rounded.
//
// The final result is rounded according to the context; it will almost always be correctly rounded,
// but may be up to 1 ulp in error in rare cases.
//
// Returns res.
func (res *Number) Power(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberPower(res.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return res
}

// NumberRescale forces exponent to a requested value. Computes res = op(lhs,rhs) where op adjusts the
// coefficient of res (by rounding or shifting) such that the exponent (-scale) of res has the value rhs.
// The numerical value of res will equal lhs, except for the effects of any rounding that occurred.
//
// Returns res.
func (res *Number) Rescale(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberRescale(res.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return res
}
