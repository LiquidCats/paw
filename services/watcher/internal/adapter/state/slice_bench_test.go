package state_test

import (
	"fmt"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/internal/adapter/state"
)

func BenchmarkPersistedState_Set(b *testing.B) {
	st := state.NewSliceState[int](1000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		st.Set(i)
	}
}

func BenchmarkPersistedState_Get(b *testing.B) {
	st := state.NewSliceState[int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = st.Get()
	}
}

func BenchmarkPersistedState_SetParallel(b *testing.B) {
	st := state.NewSliceState[int](10000)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			st.Set(i)
			i++
		}
	})
}

func BenchmarkPersistedState_GetParallel(b *testing.B) {
	st := state.NewSliceState[int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = st.Get()
		}
	})
}

func BenchmarkPersistedState_MixedReadWrite(b *testing.B) {
	st := state.NewSliceState[int](1000)

	// Pre-populate
	for i := range 100 {
		st.Set(i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				st.Set(i)
			} else {
				_ = st.Get()
			}
			i++
		}
	})
}

func BenchmarkPersistedState_Set_VaryingCapacity(b *testing.B) {
	capacities := []int{10, 100, 1000, 10000}

	for _, cap := range capacities {
		b.Run(fmt.Sprintf("capacity_%d", cap), func(b *testing.B) {
			st := state.NewSliceState[int](cap)

			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				st.Set(i)
			}
		})
	}
}

func BenchmarkPersistedState_SetWithEviction(b *testing.B) {
	st := state.NewSliceState[int](100)

	// Fill to capacity first
	for i := range 100 {
		st.Set(i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		st.Set(i) // Every set triggers eviction
	}
}
