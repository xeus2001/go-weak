package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/weak"
	"testing"
)

func TestFnv1a_a(t *testing.T) {
	// See: https://md5calc.com/hash/fnv1a64/a
	// a = af63dc4c8601ec8c
	b := []byte{'a'}
	array := weak.ToByteArray(&b)
	fnv1a := weak.Fnv1a(array, 1)
	if fnv1a != 0xaf63dc4c8601ec8c {
		t.Errorf("Expected 0xaf63dc4c8601ec8c, but found %x", fnv1a)
	}
}

func TestFnv1a_abc(t *testing.T) {
	// See: https://md5calc.com/hash/fnv1a64/a
	// abc = e71fa2190541574b
	b := []byte{'a', 'b', 'c'}
	array := weak.ToByteArray(&b)
	fnv1a := weak.Fnv1a(array, 3)
	if fnv1a != 0xe71fa2190541574b {
		t.Errorf("Expected 0xe71fa2190541574b, but found %x", fnv1a)
	}
}

func TestFnv1a_nil(t *testing.T) {
	h := weak.Fnv1a(nil, 10)
	if h != 0xCBF29CE484222325 {
		t.Errorf("Received wrong hash, expected 0xCBF29CE484222325, received: %x", h)
	}
}

func TestFnv1a_negativeLength(t *testing.T) {
	b := []byte{'a'}
	array := weak.ToByteArray(&b)
	h := weak.Fnv1a(array, -1)
	if h != 0xCBF29CE484222325 {
		t.Errorf("Received wrong hash, expected 0xCBF29CE484222325, received: %x", h)
	}
}
