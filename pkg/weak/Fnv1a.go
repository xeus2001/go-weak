package weak

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
