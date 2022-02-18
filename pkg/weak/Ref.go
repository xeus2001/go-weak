package weak

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

// Ref is a weak reference to the type T.
type Ref[T any] struct {
	hidden   uintptr // hidden pointer to T
	rescue   *T      // rescue pointer
	refCount int32   // ref counter
}

// NewRef creates a new wreak reference to the given value. This requires to provide a pointer to the value.
func NewRef[T any](pValue *T) *Ref[T] {
	r := &Ref[T]{(uintptr)(unsafe.Pointer(pValue)), nil, 1}
	var f func(p *T)
	f = func(p *T) {
		// Decrement reference counter, expect it to be zero, then no other reference is currently ongoing.
		if atomic.AddInt32(&r.refCount, -1) > 0 {
			// Race condition, Get() is concurrently trying to restore the weak reference.
			// So it is still in use, we need keep a rescue reference
			r.rescue = p
			runtime.SetFinalizer(p, f)
		}
	}
	runtime.SetFinalizer(pValue, f)
	return r
}

// Get returns either the reference T or nil, if the reference has been garbage collected.
func (r *Ref[T]) Get() *T {
	var t *T = nil
	if atomic.AddInt32(&r.refCount, 1) >= 2 {
		t = (*T)(unsafe.Pointer(r.hidden))
	}
	r.rescue = nil
	atomic.AddInt32(&r.refCount, -1)
	return t
}

// GetTest is only used to test race conditions
func (r *Ref[T]) GetTest() *T {
	var t *T = nil
	if atomic.AddInt32(&r.refCount, 1) >= 2 {
		time.Sleep(2 * time.Second) // OS put us on sleep before we can recover point!
		t = (*T)(unsafe.Pointer(r.hidden))
	}
	r.rescue = nil
	atomic.AddInt32(&r.refCount, -1)
	return t
}

// dec ref counter
// if ref counter > 0, set
