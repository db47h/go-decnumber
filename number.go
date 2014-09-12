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
