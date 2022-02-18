# go-weak

This is my weak reference implementation using generics. It was inspired by the implementation
of [ivanrad](https://github.com/ivanrad/go-weakref).

Basically the next step is to implement a concurrent weak hash map, as proposed by the following still open
[ticket #43615](https://github.com/golang/go/issues/43615).

## Warning

This code relies upon features of the garbage collector that may be changed in future go versions!

## Implementation details

The implementation is different from the one [ivanrad](https://github.com/ivanrad/go-weakref) and others have made,
because I think that one problem must be solved. When we restore the pointer while the GC is running, there will be a
write-barrier so that the restored pointer would be flagged as black as soon as it is created (when I understand the
latest version of the GC right). The problem is now a chicken egg one, we need to be sure that the uintptr is valid
before we restore it. This implementation solves this problem by ensuring in the finalizer that, when concurrently a 
**Get()** is executing, the last valid reference given to the finalizer is _rescued_ until all concurrent `Get()`'s 
are done.

This code should be thread safe, but the referee, the value being referred, may not be, so be careful!

## Usage

The usage requires go 1.18 and is quite simple:

```go
package main

import (
	"fmt"
	"github.com/xeus2001/go-weak/pkg/weak"
)

type Point struct {
	x int32
	y int32
}

func demo() {
	point := Point{10, 20}
	weakRef := weak.NewRef[Point](&point)
	p := weakRef.Get()
	if p != nil {
		fmt.Sprintf("Pointer is: %d, %d", (*p).x, (*p).y)
	}
}
```

## Test

I added a small make file and increased the test coverage to 100%. Do `make test` and then `make test-result`. Due to
the nature of one of the concurrent tests, it is possible that one branch is not entered; even while this is unlikely,
you may need to execute the test multiple times to get a 100% test coverage (at least for me, it works every time).