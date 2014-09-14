// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber

/*
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

// Status represents the status flags (exceptional conditions), and their names.
// The top byte is reserved for internal use
type Status uint32

const (
	ZeroStatus          Status = 0
	ConversionSyntax    Status = C.DEC_Conversion_syntax
	DivisionByZero      Status = C.DEC_Division_by_zero
	DivisionImpossible  Status = C.DEC_Division_impossible
	DivisionUndefined   Status = C.DEC_Division_undefined
	InsufficientStorage Status = C.DEC_Insufficient_storage // when malloc fails
	Inexact             Status = C.DEC_Inexact
	InvalidContext      Status = C.DEC_Invalid_context
	InvalidOperation    Status = C.DEC_Invalid_operation
	Overflow            Status = C.DEC_Overflow
	Clamped             Status = C.DEC_Clamped
	Rounded             Status = C.DEC_Rounded
	Subnormal           Status = C.DEC_Subnormal
	Underflow           Status = C.DEC_Underflow

	Errors      Status = C.DEC_Errors      // flags which are normally errors (result is qNaN, infinite, or 0)
	NaNs        Status = C.DEC_NaNs        // flags which cause a result to become qNaN
	Information Status = C.DEC_Information // flags which are normally for information only (finite results)
)

var statusString = map[Status]string{
	ZeroStatus:          "No status",
	ConversionSyntax:    "Conversion syntax",
	DivisionByZero:      "Division by zero",
	DivisionImpossible:  "Division impossible",
	DivisionUndefined:   "Division undefined",
	InsufficientStorage: "Insufficient storage",
	Inexact:             "Inexact",
	InvalidContext:      "Invalid context",
	InvalidOperation:    "Invalid operation",
	Overflow:            "Overflow",
	Clamped:             "Clamped",
	Rounded:             "Rounded",
	Subnormal:           "Subnormal",
	Underflow:           "Underflow",
}

// String returns a human-readable description of a status bit as a string..
// The bits set in the status field must comprise only bits defined.
// If no bits are set in the status field, the string “No status” is returned. If more than one
// bit is set, the string “Multiple status” is returned.
func (s Status) String() string {
	if str, ok := statusString[s]; ok {
		return str
	}
	return "Multiple status"
}

// ContextError represents an error condition for a Context. One can check if the last operation
// in a Context generated an error either with Context.ErrorStatus() (returns a ContextError cast as
// an error) or Context.TestStatus(Context.Errors) (returns true if an error occured).
type ContextError Status

// Error returns a string representation of the error status
func (e *ContextError) Error() string {
	return Status(*e).String()
}

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
//
// Unimplemented functions:
//
//	extern decContext  * decContextRestoreStatus(decContext *, uint32_t, uint32_t);
//	extern uint32_t      decContextSaveStatus(decContext *, uint32_t);
//	extern decContext  * decContextSetStatusFromString(decContext *, const char *);
//	extern decContext  * decContextSetStatusFromStringQuiet(decContext *, const char *);
//	extern uint32_t      decContextTestSavedStatus(uint32_t, uint32_t);
//
// TODO: test implementing Status() as returning *Status and make all Status releate functions members of Status
// This "could" improve status testingif we can cast *C.uint32_t to *uint32
//
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

// Status returns the status field of a Context
func (c *Context) Status() Status {
	// return Status(C.decContextGetStatus(&c.ctx))
	return Status(c.ctx.status)
}

// SetStatus sets one or more status bits in the status field of a decContext. Since traps are
// not supported in the Go implementation, it actually calls decContextSetStatusQuiet
//
// Normally, only library modules use this function. Applications may clear status bits with
// ClearStatus() or ZeroStatus() but should not set them (except, perhaps, for testing).
func (c *Context) SetStatus(newStatus Status) *Context {
	// C.decContextSetStatus(&c.ctx, C.uint32_t(newStatus))
	c.ctx.status |= C.uint32_t(newStatus)
	return c
}

// ClearStatus clears (sets to zero) one or more status bits in the status field of a Context.
//
// Any 1 (set) bit in the status argument will cause the corresponding bit to be cleared in the
// context status field.
func (c *Context) ClearStatus(mask Status) *Context {
	// C.decContextClearStatus(&c.ctx, C.uint32_t(status))
	c.ctx.status &^= C.uint32_t(mask)
	return c
}

// ZeroStatus is used to clear (set to zero) all the status bits in the status field of a Context.
func (c *Context) ZeroStatus() *Context {
	//	C.decContextZeroStatus(&c.ctx)
	c.ctx.status = 0
	return c
}

// TestStatus tests bits in context status and returns true if any of the tested bits are 1.
func (c *Context) TestStatus(mask Status) bool {
	// return C.decContextTestStatus(&c.ctx, C.uint32_t(mask))
	return Status(c.ctx.status)&mask != 0
}

// Func ErrorStatus() checks the Context status for any error condition
// and returns, as an error, a *ContextError if any, nil otherwise.
// Convert the return value with *err.(*decnumber.ContextError) to compare it
// against any of the Status values.
func (c *Context) ErrorStatus() error {
	if e := Status(c.ctx.status) & Errors; e != 0 {
		e := ContextError(e)
		return &e
	}
	return nil
}
