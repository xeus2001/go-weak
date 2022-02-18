package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/weak"
	"testing"
)

func TestToByteArray(t *testing.T) {
	b := []byte{0, 1, 2, 3}
	array := weak.ToByteArray(&b)
	if array == nil {
		t.Fatal("ToByteArray must not return nil")
	}
	l := len(b)
	if l != 4 {
		t.Fatalf("Received wrong length of byte slice, expected 4, got: %d", l)
	}
	for i := 0; i < l; i++ {
		expected := byte(i)
		if array[i] != expected {
			t.Fatalf("Found illegal value at index %d", i)
		}
	}
}

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
