package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/runtime"
	"github.com/xeus2001/go-weak/pkg/weak"
	"runtime"
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
	pool := weak.NewStringPool()
	bytes := []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd'}
	p := pool.ToString(bytes)
	if p == nil {
		t.Fatal("Pool must not return nil")
	}
	if *p != "Hello World" {
		t.Errorf("The returned string should have been 'Hello World', but was: '%s'", *p)
	}
	if pool.Size() != 1 {
		t.Fatalf("The pool does have a wrong size, expected 1")
	}
	t.Log("Drop reference to the string and execute GC")
	p = nil
	for i := 1; i < 6; i++ {
		runtime.GC()
		t.Logf("GC #%d", i)
		runtime.Gosched()
	}
	if pool.Size() != 0 {
		t.Errorf("Expected that the string was garbage collected, but found pool size of: %d", pool.Size())
	} else {
		t.Log("Pool freed the reference, new size is 0!")
	}
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
