// Copyright 2014 Denis Bernard. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dec_test

import (
	dec "."
	"fmt"
	"testing"
)

func TestPacked_ToNumber(t *testing.T) {
	// 6.2
	n1 := dec.Packed{
		[]byte{0x06, 0x2C},
		1,
	}
	// -7.25
	n2 := dec.Packed{
		[]byte{0x72, 0x5D},
		2,
	}
	number, err := n1.ToNumber(nil)
	if err != nil || number.String() != "6.2" {
		t.Fatalf("Error converting 6.2. Got %s (err: %v)\n", number, err)
	}
	number, err = n2.ToNumber(dec.NewNumber(12))
	if err != nil || number.String() != "-7.25" {
		t.Fatalf("Error converting -7.25. Got %s (err: %v)\n", number, err)
	}
}

func bytesToHex(b []byte) (s string) {
	for _, c := range b {
		s += fmt.Sprintf("%02X", c)
	}
	return
}

func TestPacked_FromNumber(t *testing.T) {
	var p dec.Packed

	ctx := dec.NewContext(dec.InitDecimal64, 0)
	n := dec.NewNumber(ctx.Digits()).FromString("3.14", ctx)
	if err := p.FromNumber(n); err != nil {
		t.Fatal("FromNumber() failed")
	}
	if s := bytesToHex(p.Buf); s != "314C" || p.Scale != 2 {
		t.Fatalf("scale: %d, digits: %s", p.Scale, s)
	}
	n = dec.NewNumber(ctx.Digits()).FromString("3.141", ctx)
	if err := p.FromNumber(n); err != nil {
		t.Fatal("FromNumber() failed")
	}
	if s := bytesToHex(p.Buf); s != "03141C" || p.Scale != 3 {
		t.Fatalf("scale: %d, digits: %s", p.Scale, s)
	}
	n = dec.NewNumber(ctx.Digits()).Zero()
	if err := p.FromNumber(n); err != nil {
		t.Fatal("FromNumber() failed")
	}
	if s := bytesToHex(p.Buf); s != "0C" || p.Scale != 0 {
		t.Fatalf("scale: %d, digits: %s", p.Scale, s)
	}
}
