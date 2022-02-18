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

// Alloc allocates the give amount of byte and returns an unsafe pointer to them.
func Alloc(size int) unsafe.Pointer {
	if size <= 0 {
		return nil
	}
	bytes := make([]byte, size)
	return unsafe.Pointer(ToByteArray(&bytes))
}
