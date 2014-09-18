// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#include "go-decnumber.h"
#include "decQuad.h"
#include "decimal128.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"unsafe"
)

const (
	QuadDigits = C.DECQUAD_Pmax
	QuadBytes  = C.DECQUAD_Bytes
)

// a decQuad represents a 128-bit decimal type in the IEEE 754 Standard for Floating Point Arithmetic.
//
// The Quad and Decimal128 structures are identical (except in name). Thus, Decimal128 specific functions
// have been merged in Quad's method set.
//
// Conversions to and from the Number internal format are not needed (typically the numbers are
// represented internally in “unpacked” BCD or in a base of some other power of ten), and no memory
// allocation is necessary, so Quads are much faster than using Number for arithmetic computations.
type Quad C.decQuad

// Bytes[] returns the contents of the number as a raw byte slice.
func (q *Quad) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(q), QuadBytes)
}

// FromString converts a string to a Quad.
//
// The context is supplied to this routine is used for error handling
// (setting of status and traps) and for the rounding mode, only.
// If an error occurs, the result will be a valid Quad NaN.
func (q *Quad) FromString(s string, ctx *Context) *Quad {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decQuadFromString((*C.decQuad)(q), str, ctx.DecContext())
	return q
}

// String converts a Quad to a string.
//
// No error is possible, and no status can be set.
func (q *Quad) String() string {
	str := make([]byte, C.DECQUAD_String) // TODO: escapes to heap, need to check how fmt uses sync.Pool
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decQuadToString((*C.decQuad)(q), pStr)
	return string(str[:C.strlen(pStr)])
}

// EngString converts a Quad to a string in engineering format.
//
// No error is possible, and no status can be set.
func (q *Quad) EngString() string {
	str := make([]byte, C.DECQUAD_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decQuadToEngString((*C.decQuad)(q), pStr)
	return string(str[:C.strlen(pStr)])
}

// ToNumber converts a Quad to a Number.
//
// The target number n must have appropriate space. If n is nil, a new Number will be created with
// enough storage space.
//
// No error is possible.
func (q *Quad) ToNumber(n *Number) *Number {
	if n == nil {
		n = NewNumber(QuadDigits)
	}
	C.decimal128ToNumber((*C.decimal128)(unsafe.Pointer(q)), n.DecNumber())
	return n
}

// FromNumber converts a Number to a Quad.
//
// The Context is used only for status reporting and for the rounding mode (used if the coefficient
// is more than QuadDigits digits or an overflow is detected). If the exponent is out of the
// valid range then Overflow or Underflow will be raised.  After Underflow a subnormal result is
// possible.
//
// Clamped is set if the number has to be 'folded down' to fit, by reducing its exponent and
// multiplying the coefficient by a power of ten, or if the exponent on a zero had to be
// clamped.
//
// returns q.
func (q *Quad) FromNumber(source *Number, ctx *Context) *Quad {
	C.decimal128FromNumber((*C.decimal128)(unsafe.Pointer(q)), source.DecNumber(), ctx.DecContext())
	return q
}

// Canonical copies an enoding, ensuring it is canonical.
//
// source may be the same as q.
//
// Returns q.
//
// No error is possible.
func (q *Quad) Canonical(source *Quad) *Quad {
	C.decQuadCanonical((*C.decQuad)(q), (*C.decQuad)(source))
	return q
}

// IsCanonical tests wether encoding is canonical.
func (q *Quad) IsCanonical() bool {
	return C.decQuadIsCanonical((*C.decQuad)(q)) != 0
}

func (q *Quad) Add(lhs *Quad, rhs *Quad, ctx *Context) *Quad {
	C.decQuadAdd((*C.decQuad)(q), (*C.decQuad)(lhs), (*C.decQuad)(rhs), ctx.DecContext())
	return q
}
