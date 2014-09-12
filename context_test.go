// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	dec "bitbucket.org/wildservices/go-decnumber"
	"testing"
	"unsafe"
)

func TestNewContext(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	if d := ctx.Digits(); d != 34 {
		t.Fatalf("Context init failed. Wrong number of digits. Got %d, expected 34", d)
	}
	er := dec.RoundHalfEven
	if r := ctx.Rounding(); r != er {
		t.Fatalf("Context init failed. Wrong rounding. Got %d, expected %d", r, er)
	}
}

func TestRounding(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	er := dec.RoundUp
	ctx.SetRounding(er)
	if r := ctx.Rounding(); r != er {
		t.Fatalf("Wrong rounding. Got %d, expected %d", r, er)
	}
}

func TestStatus(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	if s := ctx.Status(); s&dec.Errors != 0 || s != 0 {
		t.Fatalf("Wrong status. Got %x, expected 0", s)
	}
	es := dec.InvalidOperation
	ctx.SetStatus(es)
	if s := ctx.Status(); s&dec.Errors == 0 || s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	es = dec.Inexact
	ctx.SetStatus(es)
	ctx.ClearStatus(dec.Errors)
	if s := ctx.Status(); s&dec.Errors != 0 || s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	ctx.ZeroStatus()
	if s := ctx.Status(); s != 0 {
		t.Fatalf("Non-zero status (%x).", s)
	}
}

func TestStatusToString(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	ctx.SetStatus(dec.DivisionByZero)
	if s := ctx.StatusToString(); s != "Division by zero" {
		t.Fatalf("Wrong status to string conversion. Expected \"Division by zero\", got \"%s\"", s)
	}
}

func TestFreeNumber(t *testing.T) {
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
}
