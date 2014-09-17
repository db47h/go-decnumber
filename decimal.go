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

typedef uint8_t* decimal64b;
*/
import "C"

import "unsafe"

type Decimal64 C.decimal64

func (d *Decimal64) FromString(s string, ctx *Context) *Decimal64 {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decimal64FromString((*C.decimal64)(d), str, ctx.DecContext())
	return d
}

func (d *Decimal64) ToNumber(n *Number) *Number {
	if n == nil {
		n = NewNumber(16)
	}
	C.decimal64ToNumber((*C.decimal64)(d), n.DecNumber())
	return n
}

func (d *Decimal64) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(&d.bytes[0]), 8)
}

func (d *Decimal64) String() string {
	str := make([]byte, C.DECIMAL64_String)
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decimal64ToString((*C.decimal64)(d), pStr)
	return string(str[:C.strlen(pStr)])
}
