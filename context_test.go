// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	if d := ctx.Digits(); d != 34 {
		t.Fatalf("Context init failed. Wrong number of digits. Got %d, expected 34", d)
	}
	if e := ctx.EMax(); e != 6144 {
		t.Fatalf("Context init failed. Wrong EMax. Got %d, expected 6144", e)
	}
	if e := ctx.EMin(); e != -6143 {
		t.Fatalf("Context init failed. Wrong EMin. Got %d, expected -6143", e)
	}
	if c := ctx.Clamp(); c != 1 {
		t.Fatalf("Context init failed. Wrong Clamp. Got %d, expected 1", c)
	}
	er := dec.RoundHalfEven
	if r := ctx.Rounding(); r != er {
		t.Fatalf("Context init failed. Wrong rounding. Got %d, expected %d", r, er)
	}
}

func TestContext_Rounding(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	er := dec.RoundUp
	ctx.SetRounding(er)
	if r := ctx.Rounding(); r != er {
		t.Fatalf("Wrong rounding. Got %d, expected %d", r, er)
	}
}

// Here, we test all status methods
func TestContext_Status(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	// Status()
	// NOTE: clients should only use the Status methods, not *s
	s := ctx.Status()
	if *s != 0 {
		t.Fatalf("Wrong status. Got %x, expected 0", s)
	}
	// SetStatus() from 0
	es := dec.InvalidOperation
	s.Set(es)
	if *s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// TestStatus and dec.Errors filter
	if !s.Test(dec.Errors) {
		t.Fatal("TestStats returned false")
	}
	// check that SetStatus "adds" the requested status
	es |= dec.Overflow
	s.Set(dec.Overflow)
	if *s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// ClearStatus() only clears the requested status
	es = dec.Overflow
	s.Clear(dec.InvalidOperation)
	if *s != es {
		t.Fatalf("Wrong status. Got %x, expected %x", s, es)
	}
	// ZeroStatus()
	s.Zero()
	if *s != 0 {
		t.Fatalf("Non-zero status (%x).", s)
	}
}

func TestStatus_String(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	ctx.Status().Set(dec.DivisionByZero)
	if s := ctx.Status().String(); s != "Division by zero" {
		t.Fatalf("Wrong status to string conversion. Expected \"Division by zero\", got \"%s\"", s)
	}
	ctx.Status().Set(dec.Overflow)
	if s := ctx.Status().String(); s != "Multiple status" {
		t.Fatalf("Wrong status to string conversion. Expected \"Multiple status\", got \"%s\"", s)
	}
}

func TestContext_ErrorStatus(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	ctx.Status().Set(dec.DivisionByZero)
	err := ctx.ErrorStatus()
	if _, ok := err.(dec.ContextError); !ok || err == nil || err.Error() != "Division by zero" {
		t.Fatalf("Bad ErrorStatus(). Expected \"Division by zero\", got \"%v\"", err)
	}
}

func TestStatus_ToError(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	err := ctx.Status().Set(dec.DivisionByZero).ToError()
	if _, ok := err.(dec.ContextError); !ok || err == nil || err.Error() != "Division by zero" {
		t.Fatalf("Bad ErrorStatus(). Expected \"Division by zero\", got \"%v\"", err)
	}
}

func TestStatus_FromString(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal64, 0)
	s := ctx.Status()
	if s.Test(dec.Errors) {
		t.Fatal("Invalid context") // status should be error free just after creation
	}
	err := s.SetFromString("Division by zero")
	if err != nil || *s != dec.DivisionByZero {
		t.Fatalf("SetFromString failed. Got %x (%v)", *s, err)
	}
	err = s.SetFromString("Invalid operation")
	if err != nil || *s != dec.DivisionByZero|dec.InvalidOperation {
		t.Fatalf("SetFromString failed. Got %x (%v)", *s, err)
	}
	err = s.SetFromString("Multiple status")
	if err == nil || err.Error() != "Conversion syntax" {
		t.Fatal("SetFromString should have failed.")
	}
	err = s.SetFromString("foobar")
	if err == nil || err.Error() != "Conversion syntax" {
		t.Fatal("SetFromString should have failed.")
	}
}
