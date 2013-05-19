// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bufs implements a simple buffer cache.
//
// The intended use scheme is like:
//
//	type Foo struct {
//		buffers bufs.Buffers
//		...
//	}
//
//	// Bar can call Qux, but not the other way around (in this example).
//	const maxFooDepth = 2
//
//	func NewFoo() *Foo {
//		return &Foo{buffers: bufs.New(maxFooDepth), ...}
//	}
//
//	func (f *Foo) Bar(n int) {
//		buf := f.buffers.Alloc(n) // needed locally for computation and/or I/O
//		defer f.buffers.Free()
//		...
//		f.Qux(whatever)
//	}
//
//	func (f *Foo) Qux(n int) {
//		buf := f.buffers.Alloc(n) // needed locally for computation and/or I/O
//		defer f.buffers.Free()
//		...
//	}
//
// The whole idea behind 'bufs' is that when calling e.g. Foo.Bar N times, then
// normally, without using 'bufs', there will be 2*N (in this example) []byte
// buffers allocated.  While using 'bufs', only 2 buffers (in this example)
// will ever be created. For large N it can be a substantial difference.
//
// It's not a good idea to use Buffers to cache too big buffers. The cost of
// having a cached buffer is that the buffer is naturally not eligible for
// garbage collection.  Of course, that holds only while the Foo instance is
// reachable, in the above example.
//
// The buffer count limit is intentionally "hard" (read panicking), although
// configurable in New().  The rationale is to prevent recursive calls, using
// Alloc, to cause excessive, "static" memory consumption. Tune the limit
// carefully or do not use Buffers from within [mutually] recursive functions
// where the nesting depth is not realistically bounded to some rather small
// number.
//
// Buffers cannot guarantee improvements to you program performance. There may
// be a gain in case where they fit well. Firm grasp on what your code is
// actually doing, when and in what order is essential to proper use of
// Buffers. It's _highly_ recommended to first do profiling and memory
// profiling before even thinking about using 'bufs'. The real world example,
// and cause for this package, was a first correct, yet no optimizations done
// version of program; producing few MB of useful data while allocating 20+GB
// of memory.  Of course the garbage collector properly kicked in, yet the
// memory abuse caused ~80+% of run time to be spent memory management.  The
// program _was_ expected to be slow in its still development phase, but the
// bottleneck was guessed to be in I/O.  Actually the hard disk was waiting for
// the billions bytes being allocated and zeroed. Garbage collect on low
// memory, rinse and repeat.
//
// In the provided tests, TestFoo and TestFooBufs do the same simulated work,
// except the later uses Buffers while the former does not. Suggested test runs
// which show the differences:
//
//	$ go test -bench . -benchmem
//
//	or
//
//	$ go test -bench . -benchmem
//	$ go test -c
//	$ ./bufs.test -test.v -test.run Foo -test.memprofile mem.out -test.memprofilerate 1
//	$ go tool pprof bufs.test mem.out --alloc_space --nodefraction 0.0001 --edgefraction 0 -web
//	$ # Note: Foo vs FooBufs allocated memory is in hundreds of MBs vs 8 kB.
//
//	or
//
//	$ make demo # same as the above
//
//
// NOTE: Alloc/Free calls must be properly nested in the same way as in for
// example BeginTransaction/EndTransaction pairs. If your code can panic then
// the pairing should be enforced by deferred calls.
//
// NOTE: Buffers objects do not allocate any space until requested by Alloc,
// the mechanism works on demand only.
//
// FAQ: Why the 'bufs' package name?
//
// Package name 'bufs' was intentionally chosen instead of the perhaps more
// conventional 'buf'. There are already too many 'buf' named things in the
// code out there and that'll be a source of a lot of trouble. It's a bit
// similar situation as in the case of package "strings" (not "string").
package bufs

import (
	"errors"
)

// Buffers type represents a buffer ([]byte) cache.
type Buffers [][]byte

// New returns a newly created instance of Buffers with a maximum capacity of n
// buffers.
//
// NOTE: 'bufs.New(n)' is the same as 'make(bufs.Buffers, n)'.
func New(n int) Buffers {
	return make(Buffers, n)
}

// Alloc will return a buffer such that len(r) == n. It will firstly try to
// find an existing and unused buffer of big enough size. Only when there is no
// such, then one of the buffer slots is reallocated to a bigger size.
//
// It's okay to use append with buffers returned by Alloc. But it can cause
// allocation in that case and will again be producing load for the garbage
// collector. The best use of Alloc is for I/O buffers where the needed size of
// the buffer is figured out at some point of the code path in a 'final size'
// sense. Another real world example are compression/decompression buffers.
//
// NOTE: The buffer returned by Alloc _is not_ zeroed. That's okay for e.g.
// passing a buffer for io.Reader. If you need a zeroed buffer use Calloc.
//
// NOTE: Buffers returned from Alloc _must not_ be exposed/returned to your
// clients.  Those buffers are intended to be used strictly internally, within
// the methods of some "object".
//
// NOTE: Alloc will panic if there are no buffers (buffer slots) left.
func (p *Buffers) Alloc(n int) (r []byte) {
	b := *p
	if len(b) == 0 {
		panic(errors.New("Buffers.Alloc: out of buffers"))
	}

	var biggest, best int
	biggestI, bestI := -1, -1
	for i, v := range b {
		ln := len(v)

		if ln >= biggest {
			biggest, biggestI = ln, i
		}

		if ln >= n && (bestI < 0 || best > ln) {
			best, bestI = ln, i
		}
	}

	last := len(b) - 1
	if best >= n {
		r = b[bestI]
		b[last], b[bestI] = b[bestI], b[last]
		*p = b[:last]
		return
	}

	r = make([]byte, n)
	b[biggestI] = r
	b[last], b[biggestI] = b[biggestI], b[last]
	*p = b[:last]
	return
}

// Calloc will acquire a buffer using Alloc and then clears it to zeros. The
// zeroing goes up to n, not cap(r).
func (p *Buffers) Calloc(n int) (r []byte) {
	r = p.Alloc(n)
	for i := range r {
		r[i] = 0
	}
	return
}

// Free makes the lastly allocated by Alloc buffer free (available) again for
// Alloc.
//
// NOTE: Improper Free invocations, like in the sequence {New, Alloc, Free,
// Free}, will panic.
func (p *Buffers) Free() {
	b := *p
	b = b[:len(b)+1]
	*p = b
}

// Stats reports memory consumed by Buffers, without accounting for some
// (smallish) additional overhead.
func (p *Buffers) Stats() (bytes int) {
	b := *p
	b = b[:cap(b)]
	for _, v := range b {
		bytes += cap(v)
	}
	return
}
