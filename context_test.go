// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decnumber_test

import (
	dec "."
	"testing"
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

func TestContext_Rounding(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	er := dec.RoundUp
	ctx.SetRounding(er)
	if r := ctx.Rounding(); r != er {
		t.Fatalf("Wrong rounding. Got %d, expected %d", r, er)
	}
}

// Here, we test all status methods
func TestContext_Status(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	// Status()
	if s := ctx.Status(); s != 0 {
		t.Fatalf("Wrong status. Got %x, expected 0", s)
	}
	// SetStatus() from 0
	es := dec.InvalidOperation
	ctx.SetStatus(es)
	if s := ctx.Status(); s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// TestStatus and dec.Errors filter
	if !ctx.TestStatus(dec.Errors) {
		t.Fatal("TestStats returned false")
	}
	// check that SetStatus "adds" the requested status
	es |= dec.Overflow
	ctx.SetStatus(dec.Overflow)
	if s := ctx.Status(); s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// ClearStatus() only clears the requested status
	es = dec.Overflow
	ctx.ClearStatus(dec.InvalidOperation)
	if s := ctx.Status(); s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// ZeroStatus()
	ctx.ZeroStatus()
	if s := ctx.Status(); s != 0 {
		t.Fatalf("Non-zero status (%x).", s)
	}
}

func TestStatus_String(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	ctx.SetStatus(dec.DivisionByZero)
	if s := ctx.Status().String(); s != "Division by zero" {
		t.Fatalf("Wrong status to string conversion. Expected \"Division by zero\", got \"%s\"", s)
	}
	ctx.SetStatus(dec.Overflow)
	if s := ctx.Status().String(); s != "Multiple status" {
		t.Fatalf("Wrong status to string conversion. Expected \"Multiple status\", got \"%s\"", s)
	}
}

func TestContext_ErrorStatus(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128)
	ctx.SetStatus(dec.DivisionByZero)
	err := ctx.ErrorStatus()
	if _, ok := err.(dec.ContextError); !ok || err == nil || err.Error() != "Division by zero" {
		t.Fatalf("Bad ErrorStatus(). Expected \"Division by zero\", got \"%v\"", err)
	}
}
