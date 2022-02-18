package weak_test

import (
	"github.com/xeus2001/go-weak/pkg/weak"
	"runtime"
	"sync"
	"testing"
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
		t.Fatal("0: Reference is nil")
	}
	v := testRef.Get()
	if v == nil {
		t.Fatal("0: Value is nil")
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
	wSTART := sync.WaitGroup{}
	wSTART.Add(1)
	wEND := sync.WaitGroup{}
	wEND.Add(1)
	wDIE := sync.WaitGroup{}
	wDIE.Add(1)
	t.Log("0: Start go routing (1)")
	go func() {
		wSTART.Done()
		t.Log("1: Go routing started, do long worst case GetTest")
		v2 := testRef.GetTest()
		if v2 == nil {
			t.Errorf("1: Reference is nil, but must not happen")
		}
		t.Log("1: Reference is alive, GetTest() returned it, ok!")
		wEND.Done()
		t.Log("1: Keep the reference alive until main routing is ready")
		wDIE.Wait()
		t.Log("1: Die")
		runtime.Gosched()
	}()
	t.Log("0: Wait for go routing (1) to start")
	wSTART.Wait()
	t.Log("0: Drop all references (there are none left now!) while go method (1) hangs in GetTest()")
	runtime.Gosched()
	p = nil
	v = nil
	t.Log("0: Run GC ...")
	runtime.Gosched()
	for i := 1; i < 6; i++ {
		runtime.GC()
		t.Logf("0: GC #%d done", i)
		runtime.Gosched()
	}
	t.Log("0: Wait for the Go method to finish")
	runtime.Gosched()
	wEND.Wait()
	if testRef.Get() == nil {
		t.Fatal("0: The test reference is nil, which must not happen!")
	}
	t.Log("0: The reference is still alive, wait for the go method (1) to die")
	runtime.Gosched()
	wDIE.Done()
	runtime.Gosched()
	t.Log("0: Drop all references and start GC again, this time the weak reference should fall")
	p = nil
	v = nil
	t.Log("0: Run GC ...")
	runtime.Gosched()
	for i := 1; i < 6; i++ {
		runtime.Gosched()
		t.Logf("0: GC #%d done", i)
		runtime.GC()
	}
	p = testRef.Get()
	if p != nil {
		t.Fatal("0: p is not nil, the weak referee still lives, must not happen")
	}
	t.Log("0: The referee was collected")
}

func TestNewDeadRef(t *testing.T) {
	var _nil *point3D
	testRef := weak.NewRef[point3D](_nil)
	if testRef.Get() != nil {
		t.Errorf("Expected test reference to be nil!")
	}
}

func TestNewDeadRef2(t *testing.T) {
	var _nil *point3D
	testRef := weak.NewRef[point3D](_nil)
	if testRef.GetTest() != nil {
		t.Errorf("Expected test reference to be nil!")
	}
}
func TestConcurrentGet(t *testing.T) {
	p1 := &point3D{1, 2, 3}
	testRef := weak.NewRef[point3D](p1)
	w := sync.WaitGroup{}
	w.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			for x := 0; x < 10000; x++ {
				if testRef.Get() == nil {
					t.Fatal("Reference must not be nil!")
				}
				runtime.Gosched()
			}
			w.Done()
		}()
	}
	w.Wait()
}

func TestConcurrentGet2(t *testing.T) {
	p1 := &point3D{1, 2, 3}
	testRef := weak.NewRef[point3D](p1)
	w := sync.WaitGroup{}
	w.Add(2)
	go func() {
		if testRef.GetTest() == nil {
			t.Errorf("Reference must not be nil!")
		}
		w.Done()
	}()
	go func() {
		if testRef.GetTest() == nil {
			t.Errorf("Reference must not be nil!")
		}
		w.Done()
	}()
	w.Wait()
}
