package weak

import "unsafe"

type _byteSlice struct {
	bytes *byte
	len   int
	cap   int
}

type _string struct {
	elements *byte
	length   int
}

type ByteArray [2147483648]byte

func (a *ByteArray) Equals(other *ByteArray, length uint) bool {
	if other == nil {
		return false
	}
	for i := uint(0); i < length; i++ {
		if a[i] != other[i] {
			return false
		}
	}
	return true
}

// StringToByteArray returns a pointer to the byte array underlying the string. If the string is nil or empty, the
// method will return nil.
func StringToByteArray(string *string) *ByteArray {
	if string == nil || len(*string) == 0 {
		return nil
	}
	s := (*_string)(unsafe.Pointer(string))
	return (*ByteArray)(unsafe.Pointer(s.elements))
}

// SliceToByteArray returns the reference to the underlying byte array of the given slice. If nil is given or an empty
// byte array, then nil is returned.
func SliceToByteArray(bytes *[]byte) *ByteArray {
	if bytes == nil || len(*bytes) == 0 {
		return nil
	}
	p := (*_byteSlice)(unsafe.Pointer(bytes))
	return (*ByteArray)(unsafe.Pointer((*p).bytes))
}

// Alloc allocates the give amount of byte and returns an runtime pointer to them.
func Alloc(size int) unsafe.Pointer {
	if size <= 0 {
		return nil
	}
	bytes := make([]byte, size)
	return unsafe.Pointer(SliceToByteArray(&bytes))
}
