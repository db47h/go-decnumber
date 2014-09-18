// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#include "go-decnumber.h"
#include "decNumber.h"
#include <stdlib.h>
#include <string.h>

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
// Some functions, such as Number.Exp(), are described as mathematical functions. These have some
// restrictions: Context.EMax() must be <= Context.MaxMath, context.EMin must be >=
// -Context.MaxMath, and Context.Digits() must be <= Context.MaxMath. Non-zero operands to these
// functions must also fit within these bounds.
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
	nDigits := int(n.dn.digits)
	if nDigits == 0 {
		nDigits++
	}
	str := make([]byte, nDigits+14) // TODO: escapes to heap, need to check how fmt uses sync.Pool
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decNumberToString(n.dn, pStr)
	return string(str[:C.strlen(pStr)])
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

// A Pool represents an object that can be used as a generic pool. sync.Pool and
// util.Pool implement this interface.
type Pool interface {
	Get() interface{}
	Put(interface{})
}

// A NumberPool wraps a Pool to automatically type cast the result of Get() to a *Number.
// The *Context field is a convenience field to help in keeping track of the pool and associated
// Context with a single reference.
type NumberPool struct {
	Pool
	*Context
}

// Get returns a free *Number from the pool.
func (p *NumberPool) Get() *Number {
	return p.Pool.Get().(*Number)
}

// Put returns a *Number to the pool
// Not implemented. Uses promoted Pool.Put()
// func (p *numberPool) Put(n *Number) {
//	p.Pool.Put(n)
// }

//
// Arithmetic functions
//

// Abs is the absolute value operator. Computes n = abs(lhs)
//
// See also CopyAbs() for a quiet bitwise version of this.
//
// This has the same effect as Plus() unless lhs is negative, in which case it has the same
// effect as Minus().
//
// returns n.
func (n *Number) Abs(lhs *Number, ctx *Context) *Number {
	C.decNumberAbs(n.dn, lhs.dn, ctx.DecContext())
	return n
}

// Add adds two numbers. Computes n = lhs + rhs.
//
// Returns n.
func (n *Number) Add(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberAdd(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}

// And is the digitwise AND operator. Computes n = lhs & rhs.
//
// Logical function restrictions apply; a NaN is returned with InvalidOperation if a restriction is
// violated.
//
// Returns n.
func (n *Number) And(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberAnd(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}

// Class returns the Class ot a Number
func (n *Number) Class(ctx *Context) Class {
	return Class(C.decNumberClass(n.dn, ctx.DecContext()))
}

// Multiply multiplies one number by another. Computes n = lhs * rhs.
//
// Returns n.
func (n *Number) Multiply(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberMultiply(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}

// Divide divides one number by another. Computes n = lhs / rhs.
//
// Returns n.
func (n *Number) Divide(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberDivide(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}

// IsCanonical tests wether the encoding of a Number is canonical.
//
// Always returns true for Number's.
func (n *Number) IsCanonical() bool {
	return true
}

// IsFinite tests whether a number is finite.
//
// Returns true if the number is finite, or false otherwise (that is, it is an infinity or a NaN).
// No error is possible.
func (n *Number) IsFinite() bool {
	return n.dn.bits&C.DECSPECIAL == 0
}

// IsInfinite tests whether a number is infinite.
//
// Returns true if the number is infinite, or false otherwise (that is, it is a finite number or a
// NaN). No error is possible.
func (n *Number) IsInfinite() bool {
	return n.dn.bits&C.DECINF != 0
}

// IsNaN tests whether a number is a NaN (quiet or signaling).
func (n *Number) IsNaN() bool {
	return n.dn.bits&(C.DECNAN|C.DECSNAN) != 0
}

// IsNegative tests whether a number is negative (either minus zero, less than zero, or a NaN with a
// sign of 1).
//
// Note that for the Float types, this is called (for example) IsSigned(), and IsNegative() does not
// include zeros or NaNs.
func (n *Number) IsNegative() bool {
	return n.dn.bits&C.DECNEG != 0
}

// IsNormal tests whether a number is normal (that is, finite, non-zero, and not subnormal).
func (n *Number) IsNormal(ctx *Context) bool {
	return C.decNumberIsNormal(n.dn, ctx.DecContext()) != 0
}

// IsQNaN tests whether a number is a Quiet NaN.
func (n *Number) IsQNaN() bool {
	return n.dn.bits&C.DECNAN != 0
}

// IsSNaN tests whether a number is a Signaling NaN.
func (n *Number) IsSNaN() bool {
	return n.dn.bits&C.DECSNAN != 0
}

// IsSpecial tests whether a number has a special value (Infinity or NaN); it is the inversion of
// IsFinite()
func (n *Number) IsSpecial() bool {
	return n.dn.bits&C.DECSPECIAL != 0
}

// IsSubnormal tests whether a number is subnormal (that is, finite, non-zero, and magnitude
// less than 10^emin).
func (n *Number) IsSubnormal(ctx *Context) bool {
	return C.decNumberIsSubnormal(n.dn, ctx.DecContext()) != 0
}

// IsZero tests whether a number is a zero (either positive or negative).
func (n *Number) IsZero() bool {
	return n.dn.lsu[0] == 0 && n.dn.digits == 1 && n.dn.bits&C.DECSPECIAL == 0
}

// Radix returns the radix (number base) used by the dec package. This always returns
// 10. No error is possible.
func (n *Number) Radix() int {
	return 10
}

// Power raises a number to a power. Computes n = lhs ** rhs (lhs raised to the power of rhs).
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
// Returns n.
func (n *Number) Power(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberPower(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}

// Rescale forces exponent to a requested value. Computes n = op(lhs,rhs) where op adjusts the
// coefficient of n (by rounding or shifting) such that the exponent (-scale) of n has the value rhs.
// The numerical value of n will equal lhs, except for the effects of any rounding that occurred.
//
// Returns n.
func (n *Number) Rescale(lhs *Number, rhs *Number, ctx *Context) *Number {
	C.decNumberRescale(n.dn, lhs.dn, rhs.dn, ctx.DecContext())
	return n
}
