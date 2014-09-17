// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#cgo CFLAGS: -Ilibdecnumber

#include "go-decnumber.h"
#include "decNumber.h"
#include "decPacked.h"
*/
import "C"

// Packed represents a packed decimal number.
// Packed Decimal numbers are held as a sequence of Binary Coded Decimal digits in a byte slice,
// most significant first (at the lowest offset into the byte array) and one per 4 bits (that is,
// each digit taking a value of 0–9, and two digits per byte), with optional leading zero digits.
// The final sequence of 4 bits (called a “nibble”) will have a value greater than nine which is
// used to represent the sign of the number. The sign nibble may be any of the six possible values:
//
//	1010 (0x0a) plus
//	1011 (0x0b) minus
//	1100 (0x0c) plus (preferred)
//	1101 (0x0d) minus (preferred)
//	1110 (0x0e) plus
//	1111 (0x0f) plus
//
// Conventionally, the 0x0f sign code can also be used to indicate that a number was originally unsigned.
//
// The scale of a packed decimal number is the number of digits that follow the decimal point, and
// hence, for example, if a Packed Decimal number has the value -123456 with a scale of 2, then the
// value of the combination is -1234.56. A negative scale multiply the number by 10^(-scale).
// Combining -123456 with a scale of -2 will have the value -12345600.
type Packed struct {
	Buf   []byte
	Scale int32
}

// ToNumber converts a BCD Packed Decimal to Number.
// The BCD packed decimal byte array, together with an associated scale, is converted to a
// decNumber. The BCD array is assumed full of digits, and must be ended by a 4-bit sign nibble in
// the least significant four bits of the final byte.
//
// The scale is used (negated) as the exponent of the decNumber. Note that zeros may have a sign
// and/or a scale.
//
// If the provided Number is nil, a new Number is created with sufficient space to hold the
// converted number, so no error is possible unless the adjusted exponent is out of range, no sign
// nibble was found, or a sign nibble was found before the final nibble.  In these error cases,
// non-nil ContextError is returned and the Number will be 0.
func (p *Packed) ToNumber(num *Number) (*Number, error) {
	lp := len(p.Buf)
	if num == nil {
		sz := int32(lp)
		if sz == 0 {
			sz = 1
		}
		num = NewNumber(sz*2 - 1)
	}
	if len(p.Buf) == 0 {
		return num.Zero(), ContextError(InvalidOperation)
	}
	res := C.decPackedToNumber((*C.uint8_t)(&p.Buf[0]), C.int32_t(len(p.Buf)), (*C.int32_t)(&p.Scale), num.DecNumber())
	if res == nil {
		return num, ContextError(InvalidOperation)
	}
	return num, nil
}

// FromNumber converts a Number to BCD Packed Decimal.
//
// The Number is converted to a BCD packed decimal byte array, right aligned in the bcd array. The
// final 4-bit nibble in the array will be a sign nibble, C (1100) for + and D (1101) for -. Unused
// bytes and nibbles to the left of the number are set to 0.
//
// The PAcked scale is set to the scale of the number (this is the exponent, negated). To force the
// number to a specified scale, first use the decNumberRescale routine, which will round and change
// the exponent as necessary (TODO: but the Go implementation may fail to allocate enough space).
//
// If there is an error (that is, the Number has too many digits to fit the byte array, or it is a
// NaN or Infinity), NULL is returned and the bcd and scale results are unchanged.  Otherwise bcd is
// returned.
func (p *Packed) FromNumber(num *Number) error {
	p.Buf = make([]byte, (num.Digits()+2)/2)
	res := C.decPackedFromNumber((*C.uint8_t)(&p.Buf[0]), C.int32_t(len(p.Buf)),
		(*C.int32_t)(&p.Scale), num.DecNumber())
	if res == nil {
		return ContextError(InvalidOperation)
	}
	return nil
}
