// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"testing"
)

var (
	numberContext = dec.NewContext(dec.InitQuad, 0)
	gp            = dec.NumberPool{
		&dec.Pool{New: func() interface{} { return dec.NewNumber(numberContext.Digits()) }},
		numberContext,
	}
)

func TestNumber_String(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	n := dec.NewNumber(ctx.Digits()).FromString("1.27", ctx)
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	if s := n.String(); s != "1.27" {
		t.Fatalf("1.27 != %v\n", s)
	}
}

func TestNumber_Zero(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal128, 0)
	n := dec.NewNumber(ctx.Digits()).FromString("1.27", ctx)
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err)
	}
	s := n.Zero().String()
	if s != "0" {
		t.Fatalf("0 != %v\n", s)
	}
}

func TestNumber_Abs(t *testing.T) {
	ctx := gp.Context
	n := gp.Get().FromString("12.3", ctx)
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

func TestNumber_And(t *testing.T) {
	ctx := gp.Context
	a := gp.Get().FromString("101", ctx)
	b := gp.Get().FromString("1110", ctx)
	a.And(a, b, ctx)
	if a.String() != "100" {
		t.Fatalf("Got %s", a)
	}
}
