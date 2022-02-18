package weak

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	DEAD  = 0
	ALIVE = 1
	USE   = 2
)

// Ref is a weak reference to the type T.
type Ref[T any] struct {
	hidden uintptr // hidden pointer to T
	state  uint32  // state
}

// NewRef creates a new wreak reference to the given value. This requires to provide a pointer to the value. If the
// pointer is nil, the method will return a dead weak reference. The method never returns nil.
func NewRef[T any](pValue *T) *Ref[T] {
	var r *Ref[T]
	if pValue != nil {
		r = &Ref[T]{(uintptr)(unsafe.Pointer(pValue)), ALIVE}
		var f func(p *T)
		f = func(p *T) {
			if atomic.CompareAndSwapUint32(&r.state, ALIVE, DEAD) {
				r.hidden = 0
				return
			}
			// Concurrency: A Get() method currently uses the weak-reference. However, it is not guranteed that this
			// method did already create a real pointer (visible for the GC) from the uintptr. Therefore, we need to
			// wait until it did. In any case in this race condition, the live-time of the weak-reference should be
			// extended, so register the finalizer again.
			runtime.SetFinalizer(p, f)
			// Grab the lock on the weak reference in a spin-loop.
			for !atomic.CompareAndSwapUint32(&r.state, ALIVE, USE) {
				runtime.Gosched()
			}
			// We have the lock, that means that no other Get() method is currently using the uintptr, so
			// we can release the lock again.
			r.state = ALIVE
		}
		runtime.SetFinalizer(pValue, f)
	} else {
		r = &Ref[T]{0, DEAD}
	}
	return r
}

// Get returns either the referee T or nil, if the referee has been garbage collected.
func (r *Ref[T]) Get() *T {
repeat:
	if atomic.CompareAndSwapUint32(&r.state, ALIVE, USE) {
		// The happy path, when we can use the weak reference exclusively for us.
		t := (*T)(unsafe.Pointer(r.hidden))
		// Release our lock using a store-store write-barrier.
		atomic.StoreUint32(&r.state, ALIVE)
		// Return the reference.
		return t
	}
	if atomic.LoadUint32(&r.state) == DEAD {
		return nil
	}
	// Concurrency: Multiple threads are concurrently trying to use the reference.
	runtime.Gosched()
	goto repeat
}

// GetTest is only used to test race conditions.
func (r *Ref[T]) GetTest() *T {
repeat:
	if atomic.CompareAndSwapUint32(&r.state, ALIVE, USE) {
		// Note: This is the critical position, if the OS does switch us away here, we own the lock, but do not
		// yet have recovered the reference. If now the GC collects the underlying, the finalizer need to keep it
		// alive until we're able to restore the GC visible pointer below. We simulate this now:
		time.Sleep(3 * time.Second)
		t := (*T)(unsafe.Pointer(r.hidden))
		atomic.StoreUint32(&r.state, ALIVE)
		return t
	}
	if atomic.LoadUint32(&r.state) == DEAD {
		return nil
	}
	runtime.Gosched()
	goto repeat
}
