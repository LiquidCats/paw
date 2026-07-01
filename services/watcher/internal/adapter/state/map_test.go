package state_test

import (
	"sync"
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMapState(t *testing.T) {
	st := state.NewMapState[string, int](10)

	require.NotNil(t, st)
}

func TestMapState_Set_And_Get(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("key1", 100)
	st.Set("key2", 200)
	st.Set("key3", 300)

	val1, ok1 := st.Get("key1")
	assert.True(t, ok1)
	assert.Equal(t, 100, val1)

	val2, ok2 := st.Get("key2")
	assert.True(t, ok2)
	assert.Equal(t, 200, val2)

	val3, ok3 := st.Get("key3")
	assert.True(t, ok3)
	assert.Equal(t, 300, val3)
}

func TestMapState_Get_ReturnsNotFoundForMissingKey(t *testing.T) {
	st := state.NewMapState[string, int](10)

	val, ok := st.Get("nonexistent")

	assert.False(t, ok)
	assert.Equal(t, 0, val, "should return zero value for missing key")
}

func TestMapState_Set_OverwritesExistingKey(t *testing.T) {
	st := state.NewMapState[string, string](5)

	st.Set("key", "original")
	st.Set("key", "updated")

	val, ok := st.Get("key")

	assert.True(t, ok)
	assert.Equal(t, "updated", val)
}

func TestMapState_Del(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("key1", 100)
	st.Set("key2", 200)

	st.Del("key1")

	val1, ok1 := st.Get("key1")
	assert.False(t, ok1)
	assert.Equal(t, 0, val1)

	val2, ok2 := st.Get("key2")
	assert.True(t, ok2)
	assert.Equal(t, 200, val2)
}

func TestMapState_Del_NonexistentKey(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("key1", 100)

	// Should not panic when deleting nonexistent key
	st.Del("nonexistent")

	val, ok := st.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 100, val)
}

func TestMapState_Has(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("key1", 100)

	assert.True(t, st.Has("key1"))
	assert.False(t, st.Has("key2"))
}

func TestMapState_Has_AfterDel(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("key1", 100)
	assert.True(t, st.Has("key1"))

	st.Del("key1")
	assert.False(t, st.Has("key1"))
}

func TestMapState_ConcurrentSet(t *testing.T) {
	st := state.NewMapState[int, int](1000)

	var wg sync.WaitGroup
	numGoroutines := 10
	numSetsPerGoroutine := 50

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(base int) {
			defer wg.Done()
			for j := range numSetsPerGoroutine {
				key := base*numSetsPerGoroutine + j
				st.Set(key, key*10)
			}
		}(i)
	}
	wg.Wait()

	// Verify all keys were set
	for i := range numGoroutines * numSetsPerGoroutine {
		val, ok := st.Get(i)
		assert.True(t, ok, "key %d should exist", i)
		assert.Equal(t, i*10, val, "key %d should have correct value", i)
	}
}

func TestMapState_ConcurrentSetAndGet(_ *testing.T) {
	st := state.NewMapState[int, int](100)

	var wg sync.WaitGroup

	// Writers
	wg.Add(5)
	for i := range 5 {
		go func(base int) {
			defer wg.Done()
			for j := range 20 {
				key := base*20 + j
				st.Set(key, key)
			}
		}(i)
	}

	// Readers
	wg.Add(5)
	for range 5 {
		go func() {
			defer wg.Done()
			for j := range 20 {
				_, _ = st.Get(j) // should not panic
			}
		}()
	}

	wg.Wait()
}

func TestMapState_ConcurrentSetGetDelHas(_ *testing.T) {
	st := state.NewMapState[int, int](100)

	var wg sync.WaitGroup

	// Writers
	wg.Add(3)
	for i := range 3 {
		go func(base int) {
			defer wg.Done()
			for j := range 30 {
				st.Set(base*30+j, j)
			}
		}(i)
	}

	// Readers
	wg.Add(3)
	for range 3 {
		go func() {
			defer wg.Done()
			for j := range 30 {
				_, _ = st.Get(j)
			}
		}()
	}

	// Deleters
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			for j := range 30 {
				st.Del(j)
			}
		}()
	}

	// Has checkers
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			for j := range 30 {
				_ = st.Has(j)
			}
		}()
	}

	wg.Wait()
}

func TestMapState_WithStructType(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	st := state.NewMapState[string, TestStruct](5)

	st.Set("first", TestStruct{ID: 1, Name: "first"})
	st.Set("second", TestStruct{ID: 2, Name: "second"})

	val1, ok1 := st.Get("first")
	require.True(t, ok1)
	assert.Equal(t, TestStruct{ID: 1, Name: "first"}, val1)

	val2, ok2 := st.Get("second")
	require.True(t, ok2)
	assert.Equal(t, TestStruct{ID: 2, Name: "second"}, val2)
}

func TestMapState_WithPointerValue(t *testing.T) {
	st := state.NewMapState[string, *int](5)

	val1 := 100
	val2 := 200

	st.Set("key1", &val1)
	st.Set("key2", &val2)

	result1, ok1 := st.Get("key1")
	require.True(t, ok1)
	assert.Equal(t, 100, *result1)

	result2, ok2 := st.Get("key2")
	require.True(t, ok2)
	assert.Equal(t, 200, *result2)
}

func TestMapState_WithIntKey(t *testing.T) {
	st := state.NewMapState[int, string](5)

	st.Set(1, "one")
	st.Set(2, "two")
	st.Set(3, "three")

	val, ok := st.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "two", val)

	assert.True(t, st.Has(1))
	assert.True(t, st.Has(3))
	assert.False(t, st.Has(4))
}

func TestMapState_ZeroCapacity(t *testing.T) {
	st := state.NewMapState[string, int](0)

	require.NotNil(t, st)

	st.Set("key", 100)

	val, ok := st.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 100, val)
}

func TestMapState_Get_ZeroValueStored(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("zero", 0)

	val, ok := st.Get("zero")
	assert.True(t, ok, "key with zero value should be found")
	assert.Equal(t, 0, val)

	_, okMissing := st.Get("missing")
	assert.False(t, okMissing, "missing key should return false")
}

func TestMapState_Has_ZeroValueStored(t *testing.T) {
	st := state.NewMapState[string, int](5)

	st.Set("zero", 0)

	assert.True(t, st.Has("zero"), "Has should return true for key with zero value")
	assert.False(t, st.Has("missing"), "Has should return false for missing key")
}
