// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"testing"
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
