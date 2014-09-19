// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	"."
	"./util"
	"testing"
)

var (
	numberContext = dec.NewContext(dec.InitQuad, 0)
	gnp           = dec.NumberPool{
		&util.Pool{New: func() interface{} { return dec.NewNumber(numberContext.Digits()) }},
		numberContext,
	}
)

func TestNumber_String(t *testing.T) {
	ctx := gnp.Context
	n := gnp.Get().FromString("1.27", ctx)
	defer gnp.Put(n)
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	if s := n.String(); s != "1.27" {
		t.Fatalf("1.27 != %v\n", s)
	}
}

func TestNumber_Zero(t *testing.T) {
	ctx := gnp.Context
	n := gnp.Get().FromString("1.27", ctx)
	defer gnp.Put(n)
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	s := n.Zero().String()
	if s != "0" {
		t.Fatalf("0 != %v\n", s)
	}
}

func TestNumber_Abs(t *testing.T) {
	ctx := gnp.Context
	n := gnp.Get().FromString("12.3", ctx)
	defer gnp.Put(n)
	n.Abs(n, ctx)
	if n.String() != "12.3" {
		t.Fail()
	}
	n.FromString("-12.3", ctx)
	if n.String() != "-12.3" {
		t.Fail()
	}
	n.Abs(n, ctx)
	if n.String() != "12.3" {
		t.Fail()
	}
}

func TestNumber_Add(t *testing.T) {
	ctx := gnp.Context
	n := gnp.Get().FromString("12.3", ctx)
	m := gnp.Get().FromString("-32.02", ctx)
	r := gnp.Get().Abs(m, ctx)
	defer gnp.Putn(n, m, r)

	n.Add(n, r, ctx)
	if n.String() != "44.32" {
		t.Fatal(n)
	}
	n.Add(n, m, ctx)
	if n.String() != "12.30" {
		t.Fatal(n)
	}
	n.Add(n, m, ctx)
	if n.String() != "-19.72" {
		t.Fatal(n)
	}
}

func TestNumber_And(t *testing.T) {
	ctx := gnp.Context
	a := gnp.Get().FromString("101", ctx)
	b := gnp.Get().FromString("1110", ctx)
	defer gnp.Putn(a, b)

	a.And(a, b, ctx)
	if a.String() != "100" {
		t.Fatalf("Got %s", a)
	}
}

func TestNumber_Class(t *testing.T) {
	ctx := gnp.Context
	a := gnp.Get().Zero()
	defer gnp.Put(a)
	if c := a.Class(ctx); c != dec.ClassPosZero {
		t.Fail()
	}
	a.FromString("-INF", ctx)
	if c := a.Class(ctx); c != dec.ClassNegInf {
		t.Fail()
	}
}

func TestNumber_IsXYZ(t *testing.T) {
	ctx := gnp.Context
	n := gnp.Get()
	defer gnp.Put(n)
	n.FromString("1234", ctx.ZeroStatus())
	if !n.IsCanonical() {
		t.Fatal("Not canonical")
	}
	if !n.IsFinite() {
		t.Fatal("Not finite")
	}
	if n.IsInfinite() {
		t.Fatal("Infinite")
	}
	if n.IsNaN() || n.IsQNaN() || n.IsSNaN() {
		t.Fatal("NaN")
	}
	if n.IsNegative() {
		t.Fatal("Negative")
	}
	if !n.IsNormal(ctx) {
		t.Fatal("Not normal")
	}
	if n.IsSpecial() {
		t.Fatal("Special")
	}
	if n.IsSubnormal(ctx) {
		t.Fatal("Subnormal")
	}
	if n.IsZero() {
		t.Fatal("Zero")
	}
	n.FromString("0", ctx.ZeroStatus())
	if !n.IsZero() || n.IsNormal(ctx) {
		t.Fatal("Zero is not zero or is normal")
	}
	n.FromString("jkl", ctx.ZeroStatus())
	if !n.IsNaN() {
		t.Fatal("Not NaN")
	}
	if !n.IsQNaN() {
		t.Fatal("Not QNaN")
	}
	if n.IsSNaN() {
		t.Fatal("Is SNaN")
	}
	n.FromString("-INF", ctx.ZeroStatus())
	if !n.IsSpecial() || n.IsFinite() || !n.IsInfinite() {
		t.Fatal("Finite")
	}
	if n.IsNaN() {
		t.Fatal("NaN")
	}
	if !n.IsNegative() {
		t.Fatal("Positive")
	}
}

func TestNumber_Compare(t *testing.T) {
	var (
		ctx = gnp.Context
		n1  = gnp.Get().FromString("2.1e3", ctx)
		n2  = gnp.Get().FromString("-1.24e7", ctx)
		n3  = gnp.Get().FromString("2100", ctx)
		n4  = gnp.Get().FromString("NaN", ctx)
		n   = gnp.Get()
	)
	defer gnp.Putn(n1, n2, n3, n4, n)

	if n.Compare(n1, n2, ctx); n.IsNegative() || n.IsZero() {
		t.Fatal("<= 0")
	}
	if n.Compare(n2, n1, ctx); !n.IsNegative() {
		t.Fatal(">= 0")
	}
	if n.Compare(n1, n3, ctx); !n.IsZero() {
		t.Fatal("!=0")
	}
	if n.Compare(n1, n4, ctx); !n.IsNaN() {
		t.Fatal("!NaN")
	}
	if n.CompareTotal(n1, n4, ctx); !n.IsNegative() {
		t.Fatal(">= 0")
	}
	if n.CompareTotalMag(n1, n2, ctx); !n.IsNegative() {
		t.Fatal(">= 0")
	}
}

func TestNumber_Divide(t *testing.T) {
	var (
		ctx  = gnp.Context
		n1   = gnp.Get().FromString("355", ctx)
		n2   = gnp.Get().FromString("113", ctx)
		five = gnp.Get().FromString("-5", ctx)
		r    = gnp.Get()
	)
	defer gnp.Putn(n1, n2, five, r)
	r.Divide(n1, n2, ctx)
	r.Rescale(r, five, ctx)
	if r.String() != "3.14159" {
		t.Fatal(r)
	}
	n2.Zero()
	r.Divide(n1, n2, ctx.ZeroStatus())
	if err := ctx.ErrorStatus(); !r.IsInfinite() || err == nil || !err.(*dec.ContextError).Test(dec.DivisionByZero) {
		t.Fatal(r)
	}
}
