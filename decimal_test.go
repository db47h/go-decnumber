// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"testing"
)

func TestDecimal64_FromNumber(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal64, 0)
	d := new(dec.Decimal64)
	n := dec.NewNumber(ctx.Digits()).FromString("123.4567", ctx)
	d.FromNumber(n, ctx)
	if err := ctx.ErrorStatus(); err != nil {
		t.Fatal(err.Error())
	}
	s := d.String()
	if s != "123.4567" {
		t.Fatalf("Expected 1234.567, got %s", s)
	}
}

func TestDecimal64_EngString(t *testing.T) {
	ctx := dec.NewContext(dec.InitDecimal64, 0)
	d := new(dec.Decimal64)
	d.FromString("123.4e7", ctx)
	s := d.EngString()
	if s != "1.234E+9" {
		t.Fatalf("Expected 1.234E+9, got %s", s)
	}
}
