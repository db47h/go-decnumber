// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber

/*
#cgo CFLAGS: -Ilibdecnumber

#include "go-decnumber.h"
#include "decContext.h"
#include <stdlib.h>
*/
import "C"

// FreeListSize holds the default size of the free list of Number pointers for contexts. This
// is a tunable parameter. Set it to the desired value before creating a Context.
var FreeListSize uint32 = 128

// Rounding represents the rounding mode used by a given Context.
type Rounding uint32

const (
	RoundCeiling  Rounding = iota // round towards +infinity
	RoundUp                       // round away from 0
	RoundHalfUp                   // 0.5 rounds up
	RoundHalfEven                 // 0.5 rounds to nearest even
	RoundHalfDown                 // 0.5 rounds down
	RoundDown                     // round towards 0 (truncate)
	RoundFloor                    // round towards -infinity
	Round05Up                     // round for reround
	RoundMax                      // enum must be less than this
)

// ContextKind to use when creating a new Context with NewContext()
type ContextKind int32

const (
	InitDecimal32  ContextKind = 32
	InitDecimal64  ContextKind = 64
	InitDecimal128 ContextKind = 128
	// Synonyms
	InitSingle ContextKind = InitDecimal32
	InitDouble ContextKind = InitDecimal64
	InitQuad   ContextKind = InitDecimal128
)

// Limits for the digits, emin and emax parameters in NewCustomContext()
const (
	MaxDigits = 999999999
	MinDigits = 1
	MaxEMax   = 999999999
	MinEMax   = 0
	MaxEMin   = 0
	MinEMin   = -999999999
	MaxMath   = 999999
)

// free list of numbers
type freeNumberList struct {
	size int32 // number of digits. Needed to create new numbers of the proper size
	ch   chan *Number
}

// Get a *Number from the list or create a new one
func (l *freeNumberList) Get() *Number {
	select {
	case n := <-l.ch:
		return n
	default:
	}
	return newNumber(l.size)
}

// Put back a *Number in the free list
func (l *freeNumberList) Put(n *Number) {
	select {
	case l.ch <- n:
	default:
	}
}

// A Context wraps a decNumber context, the data structure used for providing the context
// for operations and for managing exceptional conditions.
//
// Contexts must be created using the NewContext() or NewCustomContext() functions.
//
// Most accessor and status manipulation functions (one liners) have be rewriten in pure Go in
// order to allow inlining and improve performance.
type Context struct {
	ctx C.decContext
	fn  *freeNumberList
}

// NewContext creates a new context of the requested kind.
//
// Although the native byte order should be properly detected at build time, NewContext() will
// check the runtime byte order and panic if the byte order is not set correctly. If your code panics
// on this check, please file a bug report. Providing in an invalid ContextKind will also
// cause your code to panic; this is by design.
//
// For arbitrary precision arithmetic, use NewCustomContext() instead.
//
// The Context is setup as follows, depending on the specified ContextKind:
//
// InitDecimal32 (32 bits precision):
//
//	digits = 7
//	emax = 96
//	emin = -95
//	rouning = RoundHalfEven
//	clamp = 1
//
// InitDecimal64 (64 bits precision):
//
//	digits = 16
//	emax = 384
//	emin = -383
//	rouning = RoundHalfEven
//	clamp = 1
//
// InitDecimal128 (128 bits precision):
//
//	digits = 34
//	emax = 6144
//	emin = -6143
//	rouning = RoundHalfEven
//	clamp = 1
//
func NewContext(kind ContextKind) (pContext *Context) {
	if C.decContextTestEndian(1) != 0 {
		panic("Wrong byte order for this architecture. Please file a bug report.")
	}
	if kind != InitDecimal32 && kind != InitDecimal64 && kind != InitDecimal128 {
		panic("Unsupported context kind.")
	}
	pContext = new(Context)
	C.decContextDefault(&pContext.ctx, C.int32_t(kind))
	pContext.ctx.traps = 0 // disable traps
	pContext.fn = &freeNumberList{int32(pContext.ctx.digits), make(chan *Number, FreeListSize)}
	return
}

// NewCustom context returns a new Context setup with the requested parameters.
//
// digits is used to set the precision to be used for an operation. The result of an
// operation will be rounded to this length if necessary. digits should be in [MinDigits, MaxDigits].
// The maximum supported value for digits in many arithmetic operations is MaxMath.
//
// emax is used to set the magnitude of the largest adjusted exponent that is
// permitted. The adjusted exponent is calculated as though the number were expressed in
// scientific notation (that is, except for 0, expressed with one non-zero digit before the
// decimal point).
// If the adjusted exponent for a result or conversion would be larger than emax then an
// overflow results. emax should be in [MinEMax, MaxEMax]. The maximum supported value for iemax
// in many arithmetic operations is MaxMath.
//
// emin is used to set the smallest adjusted exponent that is permitted for normal
// numbers. The adjusted exponent is calculated as though the number were expressed in
// scientific notation (that is, except for 0, expressed with one non-zero digit before the
// decimal point).
// If the adjusted exponent for a result or conversion would be smaller than emin then the
// result is subnormal. If the result is also inexact, an underflow results. The exponent of
// the smallest possible number (closest to zero) will be emin-digits+1. emin is usually set to
// -emax or to -(emax-1). emin should be in [MinEMin, MaxEMin]. The minimum supported value for
// emin in many arithmetic operations is -MaxMath.
//
// round is used to select the rounding algorithm to be used if rounding is
// necessary during an operation. It must be one of the values in the Rounding
// enumeration.
//
// clamp controls explicit exponent clamping, as is applied when a result is
// encoded in one of the compressed formats. When 0, a result exponent is limited to a
// maximum of emax and a minimum of emin (for example, the exponent of a zero result
// will be clamped to be in this range). When 1, a result exponent has the same minimum
// but is limited to a maximum of emax-(digits-1). As well as clamping zeros, this may
// cause the coefficient of a result to be padded with zeros on the right in order to bring the
// exponent within range.
// For example, if emax is +96 and digits is 7, the result 1.23E+96 would have a [sign,
// coefficient, exponent] of [0, 123, 94] if clamp were 0, but would give [0, 1230000,
// 90] if clamp were 1.
// Also when 1, clamp limits the length of NaN payloads to digits-1 (rather than digits) when
// constructing a NaN by conversion from a string.
func NewCustomContext(digits int32, emax int32, emin int32, round Rounding, clamp uint8) (pContext *Context) {
	if C.decContextTestEndian(1) != 0 {
		panic("Wrong byte order for this architecture. Please file a bug report.")
	}
	pContext = new(Context)
	c := &pContext.ctx
	C.decContextDefault(c, C.DEC_INIT_BASE)
	c.digits = C.int32_t(digits)
	c.emax = C.int32_t(emax)
	c.emin = C.int32_t(emin)
	c.round = uint32(round) // weird type for enums
	c.clamp = C.uint8_t(clamp)
	c.traps = 0 // disable traps
	pContext.fn = &freeNumberList{int32(c.digits), make(chan *Number, FreeListSize)}
	return
}

// Digits gets the working precision
func (c *Context) Digits() int32 {
	return int32(c.ctx.digits)
}

// EMin returns the Context's EMin setting
func (c *Context) EMin() int32 {
	return int32(c.ctx.emin)
}

// EMax returns the Context's EMax setting
func (c *Context) EMax() int32 {
	return int32(c.ctx.emax)
}

// Clamp returns the Context's clamping setting
func (c *Context) Clamp() int8 {
	return int8(c.ctx.clamp)
}

// Rounding gets the rounding mode
func (c *Context) Rounding() Rounding {
	// return Rounding(C.decContextGetRounding(&c.ctx))
	return Rounding(c.ctx.round)
}

// SetRounding sets the rounding mode
func (c *Context) SetRounding(newRounding Rounding) *Context {
	// C.decContextSetRounding(&c.ctx, uint32(newRounding))
	c.ctx.round = uint32(newRounding) // C enums have a Go type, not C
	return c
}

// Status returns the status of a Context
func (c *Context) Status() *Status {
	// return Status(C.decContextGetStatus(&c.ctx))
	return (*Status)(&c.ctx.status)
}

// Func ErrorStatus() checks the Context status for any error condition
// and returns, as an error, a ContextError if any, nil otherwise.
// Convert the return value with err.(decnumber.ContextError) to compare it
// against any of the Status values. This is a shorthand for Context.Status().ToError()
func (c *Context) ErrorStatus() error {
	return c.Status().ToError()
}
