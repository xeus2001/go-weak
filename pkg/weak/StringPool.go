package weak

import (
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
	hidden   uintptr // hidden pointer to the string
	elements *byte   // elements points to the bytes of the string
	length   uint    // length holds the length of the string in byte
	state    uint32  // state holds the state of this weak reference
	hash     uint64  // hash is the FNV1a hash
}

// StringSize is the size of a weak string struct in byte.
const StringSize = 40

// StringArray is an arbitrary large string array type.
type StringArray [2147483648]String

// empty is the empty string returned, when trying to intern an empty string or nil.
var empty = ""

// bytes is a helper to get simplified access to the underlying bytes of a weak string structure.
func (s *String) bytes() *ByteArray {
	return (*ByteArray)(unsafe.Pointer(s.elements))
}

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
	array := SliceToByteArray(&b)
	hash := Fnv1a(array, length)
	END := pool.headSize - 1

	headStrings := pool.headStrings()
	//oldString := pool.oldStrings()
	strings := headStrings
	index := uint32(hash & uint64(END))

	s := &strings[index]
	if s.state == ALIVE && s.length == length && s.hash == hash {
		s_bytes := s.bytes()
		if s_bytes != nil {
			//for i := uint32(0); i < length; i++ {
			//}
		}
	}
	return nil
}

// Size returns the estimated amount of stored strings.
func (pool *StringPool) Size() int {
	return 0
}

// Capacity returns the maximal amount of strings that currently can be stored.
func (pool *StringPool) Capacity() int {
	return 0
}

// TODO: Iterator
//       Remove
//       Clear
//       ?
