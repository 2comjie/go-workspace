package container

import (
	"fmt"
	"hutool/logx"
	"hutool/syncx"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSyncMap_StoreAndLoad(t *testing.T) {
	m := NewSyncMap[string, int]()

	// Test Store and Load
	m.Store("key1", 100)
	val, ok := m.Load("key1")
	if !ok {
		t.Fatal("Expected key1 to be present")
	}
	if val != 100 {
		t.Fatalf("Expected 100, got %d", val)
	}

	// Test Load non-existent key
	_, ok = m.Load("key2")
	if ok {
		t.Fatal("Expected key2 to not be present")
	}
}

func TestSyncMap_LoadOrStore(t *testing.T) {
	m := NewSyncMap[string, int]()

	// First call should store the value
	actual, loaded := m.LoadOrStore("key1", 100)
	if loaded {
		t.Fatal("Expected loaded to be false on first call")
	}
	if actual != 100 {
		t.Fatalf("Expected 100, got %d", actual)
	}

	// Second call should load the existing value
	actual, loaded = m.LoadOrStore("key1", 200)
	if !loaded {
		t.Fatal("Expected loaded to be true on second call")
	}
	if actual != 100 {
		t.Fatalf("Expected 100, got %d", actual)
	}
}

func TestSyncMap_LoadOrStoreNew(t *testing.T) {
	m := NewSyncMap[string, *TestData]()

	callCount := 0
	releaseCount := 0

	ctor := func() *TestData {
		callCount++
		return &TestData{Value: 100}
	}

	release := func(v *TestData) {
		releaseCount++
	}

	// First call should create new value
	val, loaded := m.LoadOrStoreNew("key1", ctor, release)
	if loaded {
		t.Fatal("Expected loaded to be false on first call")
	}
	if val.Value != 100 {
		t.Fatalf("Expected 100, got %d", val.Value)
	}
	if callCount != 1 {
		t.Fatalf("Expected ctor to be called once, got %d", callCount)
	}
	if releaseCount != 0 {
		t.Fatalf("Expected release not to be called, got %d", releaseCount)
	}

	// Second call should load existing value and release the new one
	val2, loaded := m.LoadOrStoreNew("key1", ctor, release)
	if !loaded {
		t.Fatal("Expected loaded to be true on second call")
	}
	if val2.Value != 100 {
		t.Fatalf("Expected 100, got %d", val2.Value)
	}
	if callCount != 1 {
		t.Fatalf("Expected ctor to be called one, got %d", callCount)
	}
	if releaseCount != 0 {
		t.Fatalf("Expected release to be called zero, got %d", releaseCount)
	}
}

func TestSyncMap_LoadOrStoreNewWithSingleFlight(t *testing.T) {
	m := NewSyncMap[string, *TestData]()

	var callCount atomic.Int32
	var releaseCount atomic.Int32

	ctor := func() *TestData {
		callCount.Add(1)
		return &TestData{Value: 100}
	}

	release := func(v *TestData) {
		releaseCount.Add(1)
	}

	// Test concurrent calls with singleflight
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			val, loaded := m.LoadOrStoreNewWithSingleFlight("sf-key1", "key1", ctor, release)
			if val.Value != 100 {
				t.Errorf("Expected 100, got %d", val.Value)
			}
			_ = loaded
		}()
	}

	wg.Wait()

	// Constructor should only be called once due to singleflight
	if callCount.Load() != 1 {
		t.Fatalf("Expected ctor to be called once, got %d", callCount.Load())
	}
	if releaseCount.Load() != 0 {
		t.Fatalf("Expected release not to be called, got %d", releaseCount.Load())
	}

	// Verify value is stored
	val, ok := m.Load("key1")
	if !ok {
		t.Fatal("Expected key1 to be present")
	}
	if val.Value != 100 {
		t.Fatalf("Expected 100, got %d", val.Value)
	}
}

func TestSyncMap_LoadAndDelete(t *testing.T) {
	m := NewSyncMap[string, int]()

	m.Store("key1", 100)

	// Test LoadAndDelete existing key
	val, loaded := m.LoadAndDelete("key1")
	if !loaded {
		t.Fatal("Expected loaded to be true")
	}
	if val != 100 {
		t.Fatalf("Expected 100, got %d", val)
	}

	// Verify key is deleted
	_, ok := m.Load("key1")
	if ok {
		t.Fatal("Expected key1 to be deleted")
	}

	// Test LoadAndDelete non-existent key
	_, loaded = m.LoadAndDelete("key2")
	if loaded {
		t.Fatal("Expected loaded to be false for non-existent key")
	}
}

func TestSyncMap_Delete(t *testing.T) {
	m := NewSyncMap[string, int]()

	m.Store("key1", 100)
	m.Delete("key1")

	_, ok := m.Load("key1")
	if ok {
		t.Fatal("Expected key1 to be deleted")
	}

	// Deleting non-existent key should not panic
	m.Delete("key2")
}

func TestSyncMap_Range(t *testing.T) {
	m := NewSyncMap[string, int]()

	m.Store("key1", 100)
	m.Store("key2", 200)
	m.Store("key3", 300)

	count := 0
	sum := 0
	m.Range(func(key string, value int) bool {
		count++
		sum += value
		return true
	})

	if count != 3 {
		t.Fatalf("Expected 3 entries, got %d", count)
	}
	if sum != 600 {
		t.Fatalf("Expected sum 600, got %d", sum)
	}

	// Test early termination
	count = 0
	m.Range(func(key string, value int) bool {
		count++
		return false // Stop after first iteration
	})

	if count != 1 {
		t.Fatalf("Expected 1 iteration, got %d", count)
	}
}

func TestSyncMap_ConcurrentAccess(t *testing.T) {
	m := NewSyncMap[int, int]()

	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				m.Store(key, key)
			}
		}(i)
	}

	wg.Wait()

	// Verify all values
	for i := 0; i < numGoroutines*numOperations; i++ {
		val, ok := m.Load(i)
		if !ok {
			t.Fatalf("Expected key %d to be present", i)
		}
		if val != i {
			t.Fatalf("Expected %d, got %d", i, val)
		}
	}
}

func TestSyncMap_DifferentTypes(t *testing.T) {
	// Test with struct type
	type Person struct {
		Name string
		Age  int
	}

	m := NewSyncMap[string, Person]()
	m.Store("john", Person{Name: "John", Age: 30})

	person, ok := m.Load("john")
	if !ok {
		t.Fatal("Expected person to be present")
	}
	if person.Name != "John" || person.Age != 30 {
		t.Fatalf("Expected John, 30, got %s, %d", person.Name, person.Age)
	}

	// Test with pointer type
	m2 := NewSyncMap[int, *Person]()
	p := &Person{Name: "Jane", Age: 25}
	m2.Store(1, p)

	person2, ok := m2.Load(1)
	if !ok {
		t.Fatal("Expected person to be present")
	}
	if person2.Name != "Jane" || person2.Age != 25 {
		t.Fatalf("Expected Jane, 25, got %s, %d", person2.Name, person2.Age)
	}
}

func TestSyncMap_NilValues(t *testing.T) {
	m := NewSyncMap[string, *TestData]()

	// Store nil value
	m.Store("key1", nil)

	val, ok := m.Load("key1")
	if !ok {
		t.Fatal("Expected key1 to be present")
	}
	if val != nil {
		t.Fatal("Expected nil value")
	}

	// Test LoadOrStore with nil
	actual, loaded := m.LoadOrStore("key2", nil)
	if loaded {
		t.Fatal("Expected loaded to be false")
	}
	if actual != nil {
		t.Fatal("Expected nil value")
	}
}

func TestSyncMap_MixedOperations(t *testing.T) {
	m := NewSyncMap[string, int]()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Concurrent stores
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			m.Store(key, id)
		}(i)
	}

	// Concurrent loads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			m.Load(key)
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			m.Delete(key)
		}(i)
	}

	wg.Wait()
}

// Test helper struct
type TestData struct {
	Value int
}

// Benchmark tests
func BenchmarkSyncMap_Store(b *testing.B) {
	m := NewSyncMap[int, int]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i, i)
			i++
		}
	})
}

func BenchmarkSyncMap_Load(b *testing.B) {
	m := NewSyncMap[int, int]()
	for i := 0; i < 10000; i++ {
		m.Store(i, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % 10000)
			i++
		}
	})
}

func BenchmarkSyncMap_LoadOrStore(b *testing.B) {
	m := NewSyncMap[int, int]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.LoadOrStore(i%1000, i)
			i++
		}
	})
}

func BenchmarkSyncMap_LoadOrStoreNew(b *testing.B) {
	m := NewSyncMap[int, *TestData]()
	ctor := func() *TestData {
		return &TestData{Value: 100}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.LoadOrStoreNew(i%1000, ctor, nil)
			i++
		}
	})
}

func BenchmarkSyncMap_LoadOrStoreNewWithSingleFlight(b *testing.B) {
	m := NewSyncMap[int, *TestData]()
	ctor := func() *TestData {
		return &TestData{Value: 100}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			sfKey := fmt.Sprintf("sf-%d", i%1000)
			m.LoadOrStoreNewWithSingleFlight(sfKey, i%1000, ctor, nil)
			i++
		}
	})
}

// Example test
func ExampleSyncMap() {
	m := NewSyncMap[string, int]()

	// Store values
	m.Store("apple", 5)
	m.Store("banana", 3)
	m.Store("orange", 7)

	// Load values
	if val, ok := m.Load("apple"); ok {
		fmt.Printf("apple: %d\n", val)
	}

	// Output:
	// apple: 5
}

func TestDeleteIf(t *testing.T) {
	mp := NewSyncMap[int, int]()
	cnt := 1000000
	for i := 0; i < cnt; i++ {
		mp.Store(i, i+1)
	}
	batch := 100
	idx := atomic.Int32{}
	syncx.WaitWork(func() {
		tmp := idx.Add(1)
		for i := 0; i < batch; i++ {
			mp.Range(func(key int, v int) bool {
				cur := int(tmp-1)*batch + i + 1
				if cur == v {
					logx.Infof("删除 %d", cur)
					mp.Delete(key)
					return false
				}
				return true
			})
			//mp.DeleteIf(func(key int, v int) bool {
			//	cur := int(tmp-1)*batch + i + 1
			//	if cur == v {
			//		logx.Infof("删除 %d", cur)
			//	}
			//	return v == cur
			//})
		}
	}, cnt/batch)

	mp.Range(func(key int, v int) bool {
		logx.Infof("err %d", v)
		return true
	})
}
