// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufs

import (
	"fmt"
	"testing"
)

func Test0(t *testing.T) {
	b := New(0)
	defer func() {
		recover()
	}()

	b.Alloc(1)
	t.Fatal("unexpected success")
}

func Test1(t *testing.T) {
	b := New(1)
	expected := false
	defer func() {
		if e := recover(); e != nil && !expected {
			t.Fatal(fmt.Errorf("%v", e))
		}
	}()

	b.Alloc(1)
	expected = true
	b.Alloc(1)
	t.Fatal("unexpected success")
}

func Test2(t *testing.T) {
	b := New(1)
	expected := false
	defer func() {
		if e := recover(); e != nil && !expected {
			t.Fatal(fmt.Errorf("%v", e))
		}
	}()

	b.Alloc(1)
	b.Free()
	b.Alloc(1)
	expected = true
	b.Alloc(1)
	t.Fatal("unexpected success")
}

func Test3(t *testing.T) {
	b := New(1)
	expected := false
	defer func() {
		if e := recover(); e != nil && !expected {
			t.Fatal(fmt.Errorf("%v", e))
		}
	}()

	b.Alloc(1)
	b.Free()
	expected = true
	b.Free()
	t.Fatal("unexpected success")
}

func Test4(t *testing.T) {
	b := New(4)
	b1 := b.Alloc(1)
	b2 := b.Alloc(2)
	b3 := b.Alloc(3)
	b4 := b.Alloc(4)

	p1 := &b1[0]
	p2 := &b2[0]
	p3 := &b3[0]
	p4 := &b4[0]

	if p1 == p2 || p1 == p3 || p1 == p4 ||
		p2 == p3 || p2 == p4 ||
		p3 == p4 {
		t.Fatal(p1, p2, p3, p4)
	}

	if len(b1) != 1 || len(b2) != 2 || len(b3) != 3 || len(b4) != 4 {
		t.Fatal(len(b1), len(b2), len(b3), len(b4))
	}

	b.Free() // 4
	b.Free() // 3
	b.Free() // 2

	x := b.Alloc(2)
	if p := &x[0]; p != p2 {
		t.Fatal(p, p2)
	}

	b.Free() // 2
	x = b.Alloc(3)
	if p := &x[0]; p != p3 {
		t.Fatal(p, p3)
	}

	b.Free() // 2
	x = b.Alloc(1)
	if p := &x[0]; p != p2 || p == p1 {
		t.Fatal(p, p2, p1)
	}

	b.Free() // 2
	x = b.Alloc(5)
	if p := &x[0]; p == p1 || p == p2 || p == p3 || p == p4 {
		t.Fatal(p)
	}
}

const (
	N       = 1e5
	bufSize = 1 << 12
)

type Foo struct {
	result []byte
}

func NewFoo() *Foo {
	return &Foo{}
}

func (f *Foo) Bar(n int) {
	buf := make([]byte, n)
	sum := 0
	for _, v := range buf {
		sum += int(v)
	}
	f.result = append(f.result, byte(sum))
	f.Qux(n)
}

func (f *Foo) Qux(n int) {
	buf := make([]byte, n)
	sum := 0
	for _, v := range buf {
		sum += int(v)
	}
	f.result = append(f.result, byte(sum))
}

type FooBufs struct {
	buffers Buffers
	result  []byte
}

const maxFooDepth = 2

func NewFooBufs() *FooBufs {
	return &FooBufs{buffers: New(maxFooDepth)}
}

func (f *FooBufs) Bar(n int) {
	buf := f.buffers.Alloc(n)
	defer f.buffers.Free()

	sum := 0
	for _, v := range buf {
		sum += int(v)
	}
	f.result = append(f.result, byte(sum))
	f.Qux(n)
}

func (f *FooBufs) Qux(n int) {
	buf := f.buffers.Alloc(n)
	defer f.buffers.Free()

	sum := 0
	for _, v := range buf {
		sum += int(v)
	}
	f.result = append(f.result, byte(sum))
}

func TestFoo(t *testing.T) {
	foo := NewFoo()
	for i := 0; i < N; i++ {
		foo.Bar(bufSize)
	}
}

func TestFooBufs(t *testing.T) {
	foo := NewFooBufs()
	for i := 0; i < N; i++ {
		foo.Bar(bufSize)
	}
	t.Log("buffers.Stats()", foo.buffers.Stats())
}

func BenchmarkFoo(b *testing.B) {
	b.SetBytes(2 * bufSize)
	foo := NewFoo()
	for i := 0; i < b.N; i++ {
		foo.Bar(bufSize)
	}
}

func BenchmarkFooBufs(b *testing.B) {
	b.SetBytes(2 * bufSize)
	foo := NewFooBufs()
	for i := 0; i < b.N; i++ {
		foo.Bar(bufSize)
	}
}
