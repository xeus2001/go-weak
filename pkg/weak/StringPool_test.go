package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/runtime"
	"github.com/xeus2001/go-weak/pkg/weak"
	"testing"
	"unsafe"
)

type _string struct {
	elements *byte
	length   int
}

func TestEmptyString(t *testing.T) {
	pool := weak.NewStringPool()
	s := pool.ToString([]byte{})
	if s == nil {
		t.Fatal("The pool must never return nil string")
	}
	if len(*s) != 0 {
		t.Fatal("The string must be empty")
	}
}

func TestStringSize(t *testing.T) {
	var ws weak.String
	size := uint(unsafe.Sizeof(ws))
	if size != weak.StringSize {
		t.Fatalf("Size of the weak.String is defined as %d, but really is: %d", weak.StringSize, size)
	}
	g := g.GetG()
	if g == nil {
		t.Fatal("Shit")
	}
}

func TestWeakness(t *testing.T) {
}

func TestOneByteStrings(t *testing.T) {
	pool := weak.NewStringPool()
	for i := 'a'; i < 'b'; i++ {
		b := byte(i)
		s := pool.ToString([]byte{b})
		if s == nil {
			t.Fatalf("The pool must not return nil: %d", i)
		}
		if len(*s) != 1 {
			t.Errorf("The string must be one byte long: %d", i)
		}
		if (*s)[0] != b {
			t.Errorf("The string does not match expected value, index %d, expected: %c, found: %c", i, (*s)[0], b)
		}
		// Restore string
		s2 := pool.ToString([]byte{b})
		if s2 == nil {
			t.Errorf("The pool must not return nil: %d", i)
		}
		if s != s2 {
			t.Errorf("The pool should have returned the same pointer (reference): %d", i)
		}
	}
}
