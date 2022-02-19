package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/weak"
	"testing"
	"unsafe"
)

type allocPoint struct {
	x, y int32
}

func TestAlloc(t *testing.T) {
	u := weak.Alloc(0)
	if u != nil {
		t.Error("Expected nil, but got a reference?")
	}

	var ap allocPoint
	size := uint(unsafe.Sizeof(ap))
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

func TestSliceToByteArray(t *testing.T) {
	b := []byte{0, 1, 2, 3}
	array := weak.SliceToByteArray(&b)
	if array == nil {
		t.Fatal("SliceToByteArray must not return nil")
	}
	l := len(b)
	if l != 4 {
		t.Errorf("Received wrong length of byte slice, expected 4, got: %d", l)
	}
	for i := 0; i < l; i++ {
		expected := byte(i)
		if array[i] != expected {
			t.Errorf("Found illegal value at index %d", i)
		}
	}

	array = weak.SliceToByteArray(nil)
	if array != nil {
		t.Error("Expected nil")
	}

	b = make([]byte, 0)
	array = weak.SliceToByteArray(&b)
	if array != nil {
		t.Error("Expected nil")
	}
}

func TestStringToByteArray(t *testing.T) {
	//       01234567890
	test := "Hello World"
	array := weak.StringToByteArray(&test)
	for i := 0; i < len(test); i++ {
		if array[i] != test[i] {
			t.Errorf("Byte #%d is not %x, but: %x", i, test[i], array[i])
		}
	}

	array = weak.StringToByteArray(nil)
	if array != nil {
		t.Error("Expected nil")
	}

	test = ""
	array = weak.StringToByteArray(&test)
	if array != nil {
		t.Error("Expected nil")
	}
}

func TestEquals(t *testing.T) {
	a := (*weak.ByteArray)(weak.Alloc(64))
	b := (*weak.ByteArray)(weak.Alloc(64))
	if !a.Equals(b, 64) {
		t.Error("Expected the empty byte array to be equal!")
	}

	as := "Hello World"
	a = weak.StringToByteArray(&as)
	bs := "Hello Other"
	b = weak.StringToByteArray(&bs)
	if !a.Equals(b, 6) {
		t.Error("Expected that the first 6 characters are both 'Hello '")
	}
	if a.Equals(b, uint(len(as))) {
		t.Error("Expected that the strings are not equal!")
	}
	if a.Equals(nil, 10) {
		t.Error("Expected that a compare against nil is always false")
	}
}
