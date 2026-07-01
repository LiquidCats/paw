package state_test

import (
	"sync"
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryState(t *testing.T) {
	st := state.NewSliceState[int](10)

	require.NotNil(t, st)
	assert.Nil(t, st.Get(), "newly created state should return nil for empty slice")
}

func TestPersistedState_Set_And_Get(t *testing.T) {
	st := state.NewSliceState[string](5)

	st.Set("value1")
	st.Set("value2")
	st.Set("value3")

	result := st.Get()

	require.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, []string{"value1", "value2", "value3"}, result)
}

func TestPersistedState_Get_ReturnsNilForEmptyState(t *testing.T) {
	st := state.NewSliceState[int](10)

	result := st.Get()

	assert.Nil(t, result)
}

func TestPersistedState_Get_ReturnsCopy(t *testing.T) {
	st := state.NewSliceState[int](10)

	st.Set(1)
	st.Set(2)

	result1 := st.Get()
	result1[0] = 999 // modify the returned slice

	result2 := st.Get()

	assert.Equal(t, 1, result2[0], "modifying returned slice should not affect internal state")
}

func TestPersistedState_Set_EnforcesCapacityLimit(t *testing.T) {
	st := state.NewSliceState[int](3)

	st.Set(1)
	st.Set(2)
	st.Set(3) // capacity reached, next Set should evict oldest
	st.Set(4)
	st.Set(5)

	result := st.Get()

	require.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, []int{3, 4, 5}, result, "oldest values should be evicted when capacity is exceeded")
}

func TestPersistedState_Set_ConcurrentAccess(t *testing.T) {
	st := state.NewSliceState[int](1000)

	var wg sync.WaitGroup
	numGoroutines := 10
	numSetsPerGoroutine := 50

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(base int) {
			defer wg.Done()
			for j := range numSetsPerGoroutine {
				st.Set(base*numSetsPerGoroutine + j)
			}
		}(i)
	}
	wg.Wait()

	result := st.Get()

	require.NotNil(t, result)
	assert.Len(t, result, numGoroutines*numSetsPerGoroutine, "all values should be set concurrently")
}

func TestPersistedState_ConcurrentSetAndGet(t *testing.T) {
	st := state.NewSliceState[int](100)

	var wg sync.WaitGroup

	// Writers
	wg.Add(5)
	for i := range 5 {
		go func(base int) {
			defer wg.Done()
			for j := range 20 {
				st.Set(base*20 + j)
			}
		}(i)
	}

	// Readers
	wg.Add(5)
	for range 5 {
		go func() {
			defer wg.Done()
			for range 20 {
				_ = st.Get() // should not panic
			}
		}()
	}

	wg.Wait()

	// Final state should be consistent
	result := st.Get()
	require.NotNil(t, result)
	assert.LessOrEqual(t, len(result), 100)
}

func TestPersistedState_WithStructType(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	st := state.NewSliceState[TestStruct](5)

	st.Set(TestStruct{ID: 1, Name: "first"})
	st.Set(TestStruct{ID: 2, Name: "second"})

	result := st.Get()

	require.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, TestStruct{ID: 1, Name: "first"}, result[0])
	assert.Equal(t, TestStruct{ID: 2, Name: "second"}, result[1])
}

func TestPersistedState_CapacityOfOne(t *testing.T) {
	st := state.NewSliceState[int](1)

	st.Set(1)

	result := st.Get()
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, []int{1}, result)

	st.Set(2) // should evict 1

	result = st.Get()
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, []int{2}, result, "with capacity 1, after second set, only the newest value should remain")
}
