# Copyright 2013 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

all:
	go fmt
	go test -i
	go test
	go build
	go vet
	go install
	make todo

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alnum:]] *.go || true
	@grep -n TODO *.go || true
	@grep -n FIXME *.go || true
	@grep -n BUG *.go || true

clean:
	rm -f bufs.test mem.out *~

demo:
	go test -bench . -benchmem
	go test -c
	./bufs.test -test.v -test.run Foo -test.memprofile mem.out \
		-test.memprofilerate 1
	go tool pprof bufs.test mem.out --alloc_space --nodefraction 0.0001 \
	       --edgefraction 0	-web
	@echo "Note: Foo vs FooBufs allocated memory is in hundreds of MBs vs 8 kB."
