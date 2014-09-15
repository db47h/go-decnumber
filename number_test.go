// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	dec "."
	"testing"
	"unsafe"
)

func TestContext_FreeNumber(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	n := ctx.NewNumber()
	p := ctx.NewNumber()
	ctx.FreeNumber(n)
	// Just make sure we get the same pointer
	if unsafe.Pointer(p) == unsafe.Pointer(n) {
		t.Fatalf("subsequent calls to NemNumber() yield same object")
	}
	q := ctx.NewNumber()
	if unsafe.Pointer(q) != unsafe.Pointer(n) {
		t.Fatalf("pointer mismatch")
	}

	// test freelist capacity. Will hang if not properly implemented
	t.Log("Testing Context freelist. Will hang if buggy...")
	dec.FreeListSize = 4
	ctx = dec.NewContext(dec.InitDecimal32)
	nums := make(chan *dec.Number, 10)
	for i := 0; i < 10; i++ {
		nums <- ctx.NewNumber()
	}
	for i := 0; i < 10; i++ {
		ctx.FreeNumber(<-nums)
	}
	t.Log("Context freelist OK")
}

func TestNumber_String(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	n := ctx.NewNumberFromString("1.27")
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	if s := n.String(); s != "1.27" {
		t.Fatalf("1.27 != %v\n", s)
	}
}

func TestNumber_Zero(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	n := ctx.NewNumberFromString("1.27")
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	s := n.Zero().String()
	if s != "0" {
		t.Fatalf("0 != %v\n", s)
	}
}
