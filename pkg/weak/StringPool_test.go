package weak_test

import (
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

func TestOneByteString(t *testing.T) {
	pool := weak.NewStringPool()
	for i := 0; i < 256; i++ {
		b := byte(i)
		s := pool.ToString([]byte{b})
		if s == nil {
			t.Fatalf("The pool must never return nil string: %d", i)
		}
		if len(*s) != 1 {
			t.Fatalf("The string must be one byte long: %d", i)
		}
		if (*s)[0] != b {
			t.Fatalf("The string does not match expected value, index %d, expected: %c, found: %c", i, (*s)[0], b)
		}
	}
}

func TestStringSize(t *testing.T) {
	var ws weak.String
	size := int(unsafe.Sizeof(ws))
	if size != weak.StringSize {
		t.Fatalf("Size of the weak.String is defined as %d, but really is: %d", weak.StringSize, size)
	}
}
