package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/weak"
	"testing"
	"unsafe"
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

type allocPoint struct {
	x, y int32
}

func TestAlloc(t *testing.T) {
	u := weak.Alloc(-1)
	if u != nil {
		t.Error("Expected nil, but got a reference?")
	}

	u = weak.Alloc(0)
	if u != nil {
		t.Error("Expected nil, but got a reference?")
	}

	var ap allocPoint
	size := int(unsafe.Sizeof(ap))
	if size != 8 {
		t.Fatalf("We expect the size to be 8, but is: %d", size)
	}
	u = weak.Alloc(size)
	if u == nil {
		t.Error("Alloc should have allocated, but got nil")
	}
	p := (*allocPoint)(u)
	p.x = 1
	p.y = 2
	t.Logf("Point: %v", *p)
}
