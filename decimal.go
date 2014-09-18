// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#include "go-decnumber.h"
#define DECNUMDIGITS 34
#include "decNumber.h"
#include "decimal32.h"
#include "decimal64.h"
#include "decimal128.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import "unsafe"

const (
	Decimal32_Digits  = 7
	Decimal64_Digits  = 16
	Decimal128_Digits = 34
)

// A Decimal32 represents a 32 bits long IEEE 754 decimal-encoded compressed decimal, which provides
// up to 7 digits of decimal precision in a compact and machine-independent form.
//
// This format is different from the IEEE 754 binary-encoded format used in Go float32.
//
// Details of the format are available at http://speleotrove.com/decimal/decbits.html
type Decimal32 C.decimal32

// Bytes[] returns the contents of the number as a raw byte slice.
func (d *Decimal32) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(&d.bytes[0]), 32/8)
}

// FromString converts a string to a Decimal32.
//
// The context is supplied to this routine is used for error handling
// (setting of status and traps) and for the rounding mode, only.
// If an error occurs, the result will be a valid decimal32 NaN.
func (d *Decimal32) FromString(s string, ctx *Context) *Decimal32 {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decimal32FromString((*C.decimal32)(d), str, ctx.DecContext())
	return d
}

// String converts a Decimal32 to a string.
//
// No error is possible, and no status can be set.
func (d *Decimal32) String() string {
	str := make([]byte, C.DECIMAL32_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal32ToString((*C.decimal32)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// EngString converts a Decimal32 to a string in engineering format.
//
// No error is possible, and no status can be set.
func (d *Decimal32) EngString() string {
	str := make([]byte, C.DECIMAL32_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal32ToEngString((*C.decimal32)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// ToNumber converts a Decimal32 to a Number.
//
// The target number n must have appropriate space. If n is nil, a new Number will be created with
// enough storage space.
//
// No error is possible.
func (d *Decimal32) ToNumber(n *Number) *Number {
	if n == nil {
		n = NewNumber(Decimal32_Digits)
	}
	C.decimal32ToNumber((*C.decimal32)(d), n.DecNumber())
	return n
}

// FromNumber converts a Number to a Decimal32.
//
// The Context is used only for status reporting and for the rounding mode (used if the coefficient
// is more than decimal32_Pmax digits or an overflow is detected).  If the exponent is out of the
// valid range then Overflow or Underflow will be raised.  After Underflow a subnormal result is
// possible.
//
// Clamped is set if the number has to be 'folded down' to fit, by reducing its exponent and
// multiplying the coefficient by a power of ten, or if the exponent on a zero had to be
// clamped.
//
// returns d.
func (d *Decimal32) FromNumber(source *Number, ctx *Context) *Decimal32 {
	C.decimal32FromNumber((*C.decimal32)(d), source.DecNumber(), ctx.DecContext())
	return d
}

// Canonical copies an enoding, ensuring it is canonical.
//
// source may be the same as d.
//
// Returns d.
//
// No error is possible.
func (d *Decimal32) Canonical(source *Decimal32) *Decimal32 {
	C.decimal32Canonical((*C.decimal32)(d), (*C.decimal32)(source))
	return d
}

// IsCanonical tests wether encoding is canonical.
func (d *Decimal32) IsCanonical() bool {
	return C.decimal32IsCanonical((*C.decimal32)(d)) != 0
}

// A Decimal64 represents a 64 bits long IEEE 754 decimal-encoded compressed decimal, which provides
// up to 16 digits of decimal precision in a compact and machine-independent form.
//
// This format is different from the IEEE 754 binary-encoded format used in Go float64.
//
// Details of the format are available at http://speleotrove.com/decimal/decbits.html
type Decimal64 C.decimal64

// Bytes[] returns the contents of the number as a raw byte slice.
func (d *Decimal64) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(&d.bytes[0]), 64/8)
}

// FromString converts a string to a Decimal64.
//
// The context is supplied to this routine is used for error handling
// (setting of status and traps) and for the rounding mode, only.
// If an error occurs, the result will be a valid decimal64 NaN.
func (d *Decimal64) FromString(s string, ctx *Context) *Decimal64 {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decimal64FromString((*C.decimal64)(d), str, ctx.DecContext())
	return d
}

// String converts a Decimal64 to a string.
//
// No error is possible, and no status can be set.
func (d *Decimal64) String() string {
	str := make([]byte, C.DECIMAL64_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal64ToString((*C.decimal64)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// EngString converts a Decimal64 to a string in engineering format.
//
// No error is possible, and no status can be set.
func (d *Decimal64) EngString() string {
	str := make([]byte, C.DECIMAL64_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal64ToEngString((*C.decimal64)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// ToNumber converts a Decimal64 to a Number.
//
// The target number n must have appropriate space. If n is nil, a new Number will be created with
// enough storage space.
//
// No error is possible.
func (d *Decimal64) ToNumber(n *Number) *Number {
	if n == nil {
		n = NewNumber(Decimal64_Digits)
	}
	C.decimal64ToNumber((*C.decimal64)(d), n.DecNumber())
	return n
}

// FromNumber converts a Number to a Decimal64.
//
// The Context is used only for status reporting and for the rounding mode (used if the coefficient
// is more than DECIMAL64_Pmax digits or an overflow is detected).  If the exponent is out of the
// valid range then Overflow or Underflow will be raised.  After Underflow a subnormal result is
// possible.
//
// Clamped is set if the number has to be 'folded down' to fit, by reducing its exponent and
// multiplying the coefficient by a power of ten, or if the exponent on a zero had to be
// clamped.
//
// returns d.
func (d *Decimal64) FromNumber(source *Number, ctx *Context) *Decimal64 {
	C.decimal64FromNumber((*C.decimal64)(d), source.DecNumber(), ctx.DecContext())
	return d
}

// Canonical copies an enoding, ensuring it is canonical.
//
// source may be the same as d.
//
// Returns d.
//
// No error is possible.
func (d *Decimal64) Canonical(source *Decimal64) *Decimal64 {
	C.decimal64Canonical((*C.decimal64)(d), (*C.decimal64)(source))
	return d
}

// IsCanonical tests wether encoding is canonical.
func (d *Decimal64) IsCanonical() bool {
	return C.decimal64IsCanonical((*C.decimal64)(d)) != 0
}

// A Decimal128 represents a 128 bits long IEEE 754 decimal-encoded compressed decimal, which provides
// up to 34 digits of decimal precision in a compact and machine-independent form.
//
// This format is different from the IEEE 754 binary-encoded format.
//
// Details of the format are available at http://speleotrove.com/decimal/decbits.html
type Decimal128 C.decimal128

// Bytes[] returns the contents of the number as a raw byte slice.
func (d *Decimal128) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(&d.bytes[0]), 128/8)
}

// FromString converts a string to a Decimal128.
//
// The context is supplied to this routine is used for error handling
// (setting of status and traps) and for the rounding mode, only.
// If an error occurs, the result will be a valid decimal128 NaN.
func (d *Decimal128) FromString(s string, ctx *Context) *Decimal128 {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decimal128FromString((*C.decimal128)(d), str, ctx.DecContext())
	return d
}

// String converts a Decimal128 to a string.
//
// No error is possible, and no status can be set.
func (d *Decimal128) String() string {
	str := make([]byte, C.DECIMAL128_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal128ToString((*C.decimal128)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// EngString converts a Decimal128 to a string in engineering format.
//
// No error is possible, and no status can be set.
func (d *Decimal128) EngString() string {
	str := make([]byte, C.DECIMAL128_String) // TODO: see sync.Pool and fmt
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal128ToEngString((*C.decimal128)(d), pStr)
	return string(str[:C.strlen(pStr)])
}

// ToNumber converts a Decimal128 to a Number.
//
// The target number n must have appropriate space. If n is nil, a new Number will be created with
// enough storage space.
//
// No error is possible.
func (d *Decimal128) ToNumber(n *Number) *Number {
	if n == nil {
		n = NewNumber(Decimal128_Digits)
	}
	C.decimal128ToNumber((*C.decimal128)(d), n.DecNumber())
	return n
}

// FromNumber converts a Number to a Decimal128.
//
// The Context is used only for status reporting and for the rounding mode (used if the coefficient
// is more than decimal128_Pmax digits or an overflow is detected).  If the exponent is out of the
// valid range then Overflow or Underflow will be raised.  After Underflow a subnormal result is
// possible.
//
// Clamped is set if the number has to be 'folded down' to fit, by reducing its exponent and
// multiplying the coefficient by a power of ten, or if the exponent on a zero had to be
// clamped.
//
// returns d.
func (d *Decimal128) FromNumber(source *Number, ctx *Context) *Decimal128 {
	C.decimal128FromNumber((*C.decimal128)(d), source.DecNumber(), ctx.DecContext())
	return d
}

// Canonical copies an enoding, ensuring it is canonical.
//
// source may be the same as d.
//
// Returns d.
//
// No error is possible.
func (d *Decimal128) Canonical(source *Decimal128) *Decimal128 {
	C.decimal128Canonical((*C.decimal128)(d), (*C.decimal128)(source))
	return d
}

// IsCanonical tests wether encoding is canonical.
func (d *Decimal128) IsCanonical() bool {
	return C.decimal128IsCanonical((*C.decimal128)(d)) != 0
}
