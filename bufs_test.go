// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufs

import (
	"fmt"
	"path"
	"runtime"
	"testing"
)

var dbg = func(s string, va ...interface{}) {
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Printf("%s:%d: ", path.Base(fn), fl)
	fmt.Printf(s, va...)
	fmt.Println()
}

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
	b := New(4)        // 0, 0, 0, 0
	b10 := b.Alloc(10) // 0, 0, 0 | 10
	if n := len(b10); n != 10 {
		t.Fatal(n)
	}

	b20 := b.Alloc(20) // 0, 0 | 20, 10
	if n := len(b20); n != 20 {
		t.Fatal(n)
	}

	b30 := b.Alloc(30) // 0 | 30, 20, 10
	if n := len(b30); n != 30 {
		t.Fatal(n)
	}

	b40 := b.Alloc(40) // | 40, 30, 20, 10
	if n := len(b40); n != 40 {
		t.Fatal(n)
	}

	p10 := &b10[0]
	p20 := &b20[0]
	p30 := &b30[0]
	p40 := &b40[0]

	if len(map[*byte]int{p10: 0, p20: 0, p30: 0, p40: 0}) != 4 {
		t.Fatal(p10, p20, p30, p40)
	}

	if len(b10) != 10 || len(b20) != 20 || len(b30) != 30 || len(b40) != 40 {
		t.Fatal(len(b10), len(b20), len(b30), len(b40))
	}

	b.Free() // 40 | 30, 20, 10
	b.Free() // 40, 30 | 20, 10
	b.Free() // 40, 30, 20 | 10

	x := b.Alloc(20) // 40, 30 | 20, 10
	if n := len(x); n != 20 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p20 {
		t.Fatal(p, p20)
	}

	b.Free()        // 40, 30, 20 | 10
	x = b.Alloc(30) // 40, 20 | 30, 10
	if n := len(x); n != 30 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p30 {
		t.Fatal(p, p30)
	}

	b.Free()        // 40, 20, 30 | 10
	x = b.Alloc(10) // 40, 30 | 20, 10
	if n := len(x); n != 10 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p20 {
		t.Fatal(p, p20)
	}

	b.Free()        // 40, 30, 20 | 10
	x = b.Alloc(50) // 30, 20 | 50, 10
	if n := len(x); n != 50 {
		t.Fatal(n)
	}

	if p := &x[0]; p == p10 || p == p20 || p == p30 || p == p40 {
		t.Fatal(p, p10, p20, p30, p40)
	}

	b.Free()        // 30, 20, 50 | 10
	x = b.Alloc(15) // 30, 50 | 20, 10
	if n := len(x); n != 15 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p20 {
		t.Fatal(p, p20)
	}

	b.Free()        // 30, 50, 20 | 10
	x = b.Alloc(25) // 50, 20, | 30, 10
	if n := len(x); n != 25 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p30 {
		t.Fatal(p, p30)
	}

	x = b.Alloc(0) // 50 | 20, 30, 10
	if n := len(x); n != 0 {
		t.Fatal(n)
	}

	x = x[:1]
	if p := &x[0]; p != p20 {
		t.Fatal(p, p20)
	}

	b.Free()       // 50, 20 | 30, 10
	x = b.Alloc(1) // 50, | 20, 30, 10
	if n := len(x); n != 1 {
		t.Fatal(n)
	}

	if p := &x[0]; p != p20 {
		t.Fatal(p, p20)
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

func TestCache(t *testing.T) {
	var c Cache
	b10 := c.Get(10) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b10), 10; g != e {
		t.Fatal(g, e)
	}

	c.Put(b10) // [10]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b9 := c.Get(9) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b9), 9; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b9[0], &b10[0]; g != e {
		t.Fatal(g, e)
	}

	c.Put(b9) // [10]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b10b := c.Get(10) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b10b), 10; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b10b[0], &b10[0]; g != e {
		t.Fatal(g, e)
	}

	c.Put(b10b) // [10]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b11 := c.Get(11) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b11), 11; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b11[0], &b10[0]; g == e {
		t.Fatal(g, e)
	}

	c.Put(b11) // [11]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b9 = c.Get(9) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b9), 9; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b9[0], &b11[0]; g != e {
		t.Fatal(g, e)
	}

	c.Put(b9) // [11]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b10b = c.Get(10) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b10b), 10; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b10b[0], &b11[0]; g != e {
		t.Fatal(g, e)
	}

	c.Put(b10b) // [11]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b11b := c.Get(11) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b11b), 11; g != e {
		t.Fatal(g, e)
	}

	if g, e := &b11b[0], &b11[0]; g != e {
		t.Fatal(g, e)
	}

	c.Put(b11b) // [11]
	if g, e := len(c), 1; g != e {
		t.Fatal(g, e)
	}

	b12 := c.Get(12) // []
	if g, e := len(c), 0; g != e {
		t.Fatal(g, e)
	}

	if g, e := len(b12), 12; g != e {
		t.Fatal(g, e)
	}
}
