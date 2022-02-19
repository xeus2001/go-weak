package weak

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

//
// Concept:
// We create one linear array that holds all weak references to the strings. We try to insert each string up to
// 8 times, linear from the index that it's hash refers to. If the current table is full, we need to resize the table.
// Resize is basically simple, we first create a new table, twice the size and flag all weak strings as SEE_OLD.
// Then we copy head to old and the new table to head. Now we have the old table that is full and a new table, that
// is currently holding references to old via the state SEE_OLD.
// All readers should always enter the head table and when they hit an SEE_OLD value, they need to enter the old table
// to review the old value.
// So far the concept is quite simple, we only need to ensure that only one thread is doing the resize. We lock the new
// and old values we copy, then copy the old value into the new table, set a tombstone and unlock both. Generally the
// table should stay read and writable while the copy is ongoing, even while it is slightly slower.
// Issues will arise when multiple threads are going to require the resize, which will in the current first version
// only be allowed for one thread (later we can allow multiple threads to help to do the copy). We need additionally
// to handle the situation when concurrently to a resize a weak reference is collected, but we can do this in the
// finalizer by simply let him wait for the resize to finish.
//
// There are various points we can optimize later, for example shrink the table, but let's first get started.
//

// StringPool is a concurrent weak string pool to cache strings for parsers and alike.
type StringPool struct {
	old      *String // old is only used when the hash map is resized.
	oldSize  uint    // oldSize holds the size of the old hash map.
	head     *String // head always refer to the current head data.
	headSize uint    // newSize holds the size of the new hash map.
	state    uint32  // state holds the state of the pool.
}

// String is a weak string struct.
type String struct {
	string      uintptr    // string pointer to the string
	bytes       *ByteArray // elements points to the bytes of the string
	length      uint       // length holds the length of the string in byte
	stateAndUid uint64     // stateAndUid holds the state of this weak reference and the UID
	hash        uint64     // hash is the FNV1a hash
}

// nextUid is a pointer to the next `uid` to be used.
var nextUid *uint64 = (*uint64)(Alloc(8))

// StringSize is the size of a weak string struct in byte.
const (
	StringSize uint   = 40
	STATE_MASK uint64 = 0xff
)

// TrySlotsBeforeResize holds the amount of slots to be tested before the map is resized.
var TrySlotsBeforeResize uint = 8

// StringArray is an arbitrary large string array type.
type StringArray [2147483648]String

// empty is the empty string returned, when trying to intern an empty string or nil.
var empty = ""

// NewStringPool allocates a new string pool and returns a pointer to it. The method will never return nil.
func NewStringPool() *StringPool {
	pool := &StringPool{}
	pool.headSize = 64
	pool.head = (*String)(Alloc(64 * StringSize))
	pool.state = ALIVE
	return pool
}

// oldStrings is a helper that returns the old strings as array for simplified access.
func (pool *StringPool) oldStrings() *StringArray {
	return (*StringArray)(unsafe.Pointer(pool.old))
}

// headStrings is a helper that returns the head strings as array for simplified access.
func (pool *StringPool) headStrings() *StringArray {
	return (*StringArray)(unsafe.Pointer(pool.head))
}

// Contains tests if the given string is part of the pool.
func (pool *StringPool) Contains(s *string) bool {
	return false
}

// Intern will intern the given string and return the interned version.
func (pool *StringPool) Intern(s *string) *string {
	return nil
}

// ToString will pool the given bytes and return a string. If the bytes exist already as string, the existing string
// is returned.
func (pool *StringPool) ToString(b []byte) *string {
	length := uint(len(b))
	if length == 0 {
		return &empty
	}
	bytes := SliceToByteArray(&b)
	hash := Fnv1a(bytes, length)

	END := pool.headSize - 1
	headStrings := pool.headStrings()
	//oldString := pool.oldStrings()
	strings := headStrings
	index := uint(hash & uint64(END))

	// TODO: What happens when two threads concurrently insert the same bytes, both may not find them and then
	//       insert them. My only solution is that after we've inserted our version, we need to verify that nobody
	//       else has inserted the same string in front of us, if he has, we can simply release our version. We
	//       simply say, the version that is the closest to the original index should stay, duplicates need to release.

	// Search existing
	for search := TrySlotsBeforeResize; search > 0; search-- {
		s := &strings[index]
	spin:
		stateAndUid := s.stateAndUid
		//uid := stateAndUid >> 8
		state := uint32(s.stateAndUid & 0xff)
		if state == ALIVE && s.length == length && s.hash == hash {
			if s.bytes != nil {
				if bytes.Equals(s.bytes, length) {
					useStateAndUid := (stateAndUid & 0xffff_ffff_ffff_ff00) | uint64(USE)
					if atomic.CompareAndSwapUint64(&s.stateAndUid, stateAndUid, useStateAndUid) {
						found := (*string)(unsafe.Pointer(s.string))
						s.stateAndUid = useStateAndUid
						return found
					}
					// Concurrency: We have found the string, but it is concurrently accessed, we need to wait
					// what happens to it, it may be possible that it is garbage collected, or it is moved in memory
					// due to a resize, in any case, we need to wait.
					runtime.Gosched()
					goto spin
				}
			}
		}
		index = (index + 1) & END
	}

	// Insert string
	index = uint(hash & uint64(END))
	for insert := TrySlotsBeforeResize; insert > 0; insert-- {
		s := &strings[index]
		stateAndUid := s.stateAndUid
		//uid := stateAndUid >> 8
		state := uint32(s.stateAndUid & 0xff)
		if state == DEAD {
			useStateAndUid := (stateAndUid & 0xffff_ffff_ffff_ff00) | uint64(USE)
			if atomic.CompareAndSwapUint64(&s.stateAndUid, stateAndUid, useStateAndUid) {
				uid := atomic.AddUint64(nextUid, 1)
				str := string(b)
				s.string = (uintptr)(unsafe.Pointer(&str))
				s.bytes = bytes
				s.hash = hash
				s.length = length
				var finalizer func(str *string)
				finalizer = func(str *string) {
					alive := (s.stateAndUid & 0xffff_ffff_ffff_ff00) | uint64(ALIVE)
					if atomic.CompareAndSwapUint64(&s.stateAndUid, alive, uint64(DEAD)) {
						s.string = 0
						s.bytes = nil
						s.hash = 0
						s.length = 0
						return
					}
					// Concurrency: Another method currently uses the weak-reference. However, it is not guaranteed
					// that this method did already create a real pointer (visible for the GC) from the uintptr and
					// this time there are other concurrent things that can happen. Therefore, we need to wait until
					// we can acquire the lock to rescue the reference.
					runtime.SetFinalizer(&str, finalizer)
					// Grab the lock on the weak reference in a spin-loop.
				finSpin:
					current := atomic.LoadUint64(&s.stateAndUid)
					use := (current & 0xffff_ffff_ffff_ff00) | uint64(USE)
					if !atomic.CompareAndSwapUint64(&s.stateAndUid, current, use) {
						runtime.Gosched()
						goto finSpin
					}
					alive = (s.stateAndUid & 0xffff_ffff_ffff_ff00) | uint64(ALIVE)
					s.stateAndUid = alive
				}
				runtime.SetFinalizer(&str, finalizer)
				atomic.StoreUint64(&s.stateAndUid, (uid<<8)|uint64(ALIVE))
				return &str
			}
		}
		index = (index + 1) & END
	}
	// TODO: This operation should never fail, but in this state it will return nil as error!
	return nil
}

// Size returns the estimated amount of stored strings.
func (pool *StringPool) Size() uint {

	END := pool.headSize - 1
	headStrings := pool.headStrings()
	//oldString := pool.oldStrings()
	strings := headStrings
	var size uint = 0
	for i := uint(0); i <= END; i++ {
		if (strings[i].stateAndUid & 0xff) == uint64(ALIVE) {
			size++
		}
	}
	return size
}

// Capacity returns the maximal amount of strings that currently can be stored.
func (pool *StringPool) Capacity() int {
	return 0
}

// TODO: Iterator
//       Remove
//       Clear
//       ?
