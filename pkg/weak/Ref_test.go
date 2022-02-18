package weak_test

import (
	"go-weak/pkg/weak"
	"runtime"
	"sync"
	"testing"
	"time"
)

type point3D struct {
	x, y, z int
}

func TestNewRef(t *testing.T) {
	testRef := weak.NewRef[point3D](&point3D{1, 2, 3})
	if testRef == nil {
		t.Fatal("Reference is nil")
	}
	p := testRef.Get()
	if p == nil {
		t.Fatal("Value is nil")
	}
	if (*p).x != 1 {
		t.Errorf("Invalid x: %d", (*p).x)
	}
	if (*p).y != 2 {
		t.Errorf("Invalid y: %d", (*p).y)
	}
	if (*p).z != 3 {
		t.Errorf("Invalid z: %d", (*p).z)
	}
	p = nil
	for i := 0; i < 10; i++ {
		runtime.Gosched()
		runtime.GC()
	}
	if testRef == nil {
		t.Fatal("Reference is nil")
	}
	if testRef.Get() != nil {
		t.Fatal("Value is not nil")
	}
}

// TestRaceCondition will test what happens when the OS does interrupt out Get() method in the worst possible moment
func TestRaceCondition(t *testing.T) {
	p := &point3D{1, 2, 3}
	testRef := weak.NewRef[point3D](p)
	if testRef == nil {
		t.Fatal("Reference is nil")
	}
	v := testRef.Get()
	if v == nil {
		t.Fatal("Value is nil")
	}
	if (*v).x != (*p).x {
		t.Errorf("Invalid x: %d", (*p).x)
	}
	if (*v).y != (*p).y {
		t.Errorf("Invalid y: %d", (*p).y)
	}
	if (*v).z != (*p).z {
		t.Errorf("Invalid z: %d", (*p).z)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	t.Log("Start go routing")
	go func() {
		t.Log("Go routing started, do GetTest")
		v2 := testRef.GetTest()
		if v2 == nil {
			t.Errorf("Reference is nil, but must not happen")
		}
		t.Log("Reference was alive, happy!")
		wg.Done()
	}()
	// We still have a reference, ensure that the go func stated
	time.Sleep(1 * time.Second)
	// Drop the reference, current our "Go routing should hang"
	t.Log("Drop references and start GC")
	p = nil
	v = nil
	for i := 0; i < 10; i++ {
		runtime.Gosched()
		runtime.GC()
	}
	p = testRef.Get()
	if p == nil {
		t.Error("p is nil")
	}
	wg.Wait()
	p = nil
	v = nil
	t.Log("Drop references again and start GC again, this time the weak ref should fall")
	for i := 0; i < 10; i++ {
		runtime.Gosched()
		runtime.GC()
	}
	p = testRef.Get()
	if p != nil {
		t.Error("p is not nil")
	}
}
