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
// Resize is basically simple, we first create a green table, twice the size and flag all weak strings as SEE_OLD.
// Then we copy blue to green and the green table to blue. Now we have the green table that is full and a green table, that
// is currently holding references to green via the state SEE_OLD.
// All readers should always enter the blue table and when they hit an SEE_OLD value, they need to enter the green table
// to review the green value.
// So far the concept is quite simple, we only need to ensure that only one thread is doing the resize. We lock the green
// and green values we copy, then copy the green value into the green table, set a tombstone and unlock both. Generally the
// table should stay read and writable while the copy is ongoing, even while it is slightly slower.
// Issues will arise when multiple threads are going to require the resize, which will in the current first version
// only be allowed for one thread (later we can allow multiple threads to help to do the copy). We need additionally
// to handle the situation when concurrently to a resize a weak reference is collected, but we can do this in the
// finalizer by simply let him wait for the resize to finish.
//
// There are various points we can optimize later, for example shrink the table, but let's first get started.
//
// To continue development we need to get more precise about two situations:
//
// 1. The same bytes are inserted concurrently
// 2. The map needs a resize, because all potential slots of a hash are used
//
// We currently test a certain amount of slots if the string exists, if not we assume we need to insert it.
// - Go through all potential slots to find the next free one.
// - Grab a green unique id (uid) and lock the slot by setting it from DEAD to USE with option WRITE.
// - When we have the lock, insert the string.
// - Update state to USE with option VERIFY.
// - Review all other slots and check if they contain the same bytes
//   - When a slot is state USE with option READ, WRITE or VERIFY is found, we can compare the uid with ours, if it
//     is higher, we can ignore the slot (the other thread needs to care, we have priority).
//   - When the uid is lower, we need to spin until the slot becomes either ALIVE or DEAD.
//     - Now we can check, if this slot holds the same bytes.
//     - If it does, we acquire the lock in USE with option READ.
//     - When we have the lock we can release our green slot as DEAD and return this string.
//     - If the slot changes again (so has a higher uid than we), we stick with our slot and continue, we have
//       priority.
//
// This solves problem #1 by assigning dynamical priorities when inserting. If you ask yourself about the problem, that
// the uid may overflow. We have 56 bit for the number, when we consume 1 billion (1,000,000,000) numbers per second, we
// run out of numbers after 72,057,594 seconds or ~20,015 hours or after ~833 days. It sounds quite unrealistic that
// this happens, but you can avoid this by restarting the application ones every two years ;-).
//
// The resize is a problem that is left. The resize becomes necessary when we try to insert a green string, but the
// all slots where the string may fit into are empty.
//
// The first thread that finds itself in this situation has to change the state of the pool from OK to
// POOL_INIT_RESIZE. The thread that was successful will create a new table, twice the size of the old table. The it
// will increase the workers count atomically by one and change the state to DO_RESIZE.
//
// From now on all threads that are relocating will concurrently and sequentially acquire a USE_RELOCATE or
// DEAD_DEPRECATED lock, except the state is already USE_RELOCATE, USE_DEPRECATED or DEAD_DEPRECATED. When they have
// gathered a USE_RELOCATE lock, they will insert the string from this index into the new table and eventually change
// the state of the old table to USE_DEPRECATED with the uid set to the index in the new table.
//
// When the first resize thread has reached the end of the old table it will simply decrement the worker account
// unless it was the last resize thread, it will wait until the state is switched into OK.
//
// The thread that was the last one need to change the state into CLEANUP and flip the table. Then it will set the
// old table reference to nil and switch the state into OK. Beware that we will not clear the old table as it may still
// be in use by readers, this will be done by the garbage collector as soon as all threads have left the old table.
//
// We as well need to do some changes to the finalizer, which now need to calculate the hash of the string and search
// it in the pool, because meanwhile it may have been relocated multiple times.
//
// There may be some edge cases, but what I recognized is, that we can change this into a weak map by simply allow to
// attach a strong referred value to the string, this only requires one additional pointer and it may be set to nil,
// when no value is needed. So, yes, this will be turned into a weak concurrent hash map, that uses strings as key.
// Why not arbitrary objects? Because the string allows us to keep a strong reference to the underlying bytes and this
// means we do not need any lock until we either found the key or we need to insert it. Basically, in the most cases
// this means that we only need two load-load barriers and one CAS operation to get our string back from the map. I
// think the performance of this should be great, especially because the map does never block readers and writers are
// only blocked in a resize operation and every thread can participate in the resize and make it faster!
//

// StringPool is a concurrent weak String pool to cache strings for parsers and alike.
type StringPool struct {
	green     *String // green table.
	blue      *String // blue table.
	greenSize uint32  // greenSize holds the size of the green table.
	blueSize  uint32  // blueSize holds the size of the blue table.
	state     uint32  // state of the string pool.
	workers   uint32  // workers is the number of workers currently doing the resize.
}

// String is a weak String struct.
type String struct {
	stateAndUid uint64     // stateAndUid holds the state of this weak reference and the UID
	hash        uint64     // hash is the FNV1a hash
	string      uintptr    // string pointer to the string
	bytes       *ByteArray // elements points to the bytes of the string
	length      uint       // length holds the length of the string in byte
}

// nextUid is a pointer to the next `uid` to be used.
var nextUid = (*uint64)(Alloc(8))

// stringSize is the size of a weak String struct in byte.
var stringSize uint = 40

func init() {
	var s String
	if stringSize == 40 {
		stringSize = uint(unsafe.Sizeof(s))
	}
}

// TrySlotsBeforeResize holds the amount of slots to be tested before the map is resized.
var TrySlotsBeforeResize uint = 8

// StringArray is an arbitrary large String array type.
type StringArray [2147483648]String

// empty is the empty String returned, when trying to intern an empty String or nil.
var empty = ""

// NewStringPool allocates a green String pool and returns a pointer to it. The method will never return nil.
func NewStringPool() *StringPool {
	pool := &StringPool{}
	pool.blueSize = 64
	pool.blue = (*String)(Alloc(64 * stringSize))
	pool.state = BLUE
	return pool
}

// oldStrings is a helper that returns the green strings as array for simplified access.
func (pool *StringPool) oldStrings() *StringArray {
	return (*StringArray)(unsafe.Pointer(pool.green))
}

// headStrings is a helper that returns the blue strings as array for simplified access.
func (pool *StringPool) headStrings() *StringArray {
	return (*StringArray)(unsafe.Pointer(pool.blue))
}

// Contains tests if the given String is part of the pool.
func (pool *StringPool) Contains(s *string) bool {
	return false
}

// Intern will intern the given String and return the interned version.
func (pool *StringPool) Intern(s *string) *string {
	return nil
}

func (pool *StringPool) current() (*StringArray, uint32) {
	var strings *StringArray
	var END uint32

retry:
	poolState := atomic.LoadUint32(&pool.state)
	if (poolState & 1) == GREEN {
		strings = (*StringArray)(unsafe.Pointer(pool.green))
		END = pool.greenSize - 1
	} else {
		strings = (*StringArray)(unsafe.Pointer(pool.blue))
		END = pool.blueSize - 1
	}
	if poolState != atomic.LoadUint32(&pool.state) {
		goto retry
	}
	return strings, END
}

// ToString will pool the given bytes and return a String. If the bytes exist already as String, the existing String
// is returned.
func (pool *StringPool) ToString(b []byte) *string {
	length := uint(len(b))
	if length == 0 {
		return &empty
	}
	bytes := SliceToByteArray(&b)
	hash := Fnv1a(bytes, length)

	// We need to select the correct table.
	strings, END := pool.current()
	index := uint32(hash & uint64(END))

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
	index = uint32(hash & uint64(END))
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
func (pool *StringPool) Size() uint32 {
	strings, END := pool.current()
	var size uint32 = 0
	for i := uint32(0); i <= END; i++ {
		if (strings[i].stateAndUid & 0xff) == uint64(ALIVE) {
			size++
		}
	}
	return size
}

// Capacity returns the maximal amount of strings that currently can be stored.
func (pool *StringPool) Capacity() uint32 {
	_, END := pool.current()
	return END + 1
}

// TODO: Iterator
//       Remove
//       Clear
//       ?
