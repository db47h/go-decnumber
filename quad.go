// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec

/*
#include "go-decnumber.h"
#include "decQuad.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"unsafe"
)

type Quad C.decQuad

func (q *Quad) FromString(s string, ctx *Context) *Quad {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.decQuadFromString((*C.decQuad)(q), str, ctx.DecContext())
	return q
}

func (q *Quad) String() string {
	str := make([]byte, C.DECQUAD_String) // TODO: escapes to heap, need to check how fmt uses sync.Pool
	pStr := (*C.char)(unsafe.Pointer(&str[0]))
	C.decQuadToString((*C.decQuad)(q), pStr)
	return string(str[:C.strlen(pStr)])
}

func (q *Quad) Add(lhs *Quad, rhs *Quad, ctx *Context) *Quad {
	C.decQuadAdd((*C.decQuad)(q), (*C.decQuad)(lhs), (*C.decQuad)(rhs), ctx.DecContext())
	return q
}
