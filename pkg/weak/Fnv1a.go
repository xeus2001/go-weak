package weak

import "unsafe"

type ByteArray [2147483648]byte

type _byteSlice struct {
	bytes *byte
	len   int
	cap   int
}

// ToByteArray returns the reference to the underlying byte array of the given slice.
func ToByteArray(bytes *[]byte) *ByteArray {
	p := (*_byteSlice)(unsafe.Pointer(bytes))
	return (*ByteArray)(unsafe.Pointer((*p).bytes))
}

// Fnv1a calculates the FNV1a hash value above the given bytes.
func Fnv1a(array *ByteArray, len int) uint64 {
	var h uint64 = 0xCBF29CE484222325
	if array == nil || len <= 0 {
		return h
	}
	for i := 0; i < len; i++ {
		v := uint64(array[i])
		h ^= v & 0xFF
		h *= 1099511628211
	}
	return h
}
