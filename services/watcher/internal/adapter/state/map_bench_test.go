package state_test

import (
	"fmt"
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/state"
)

func BenchmarkMapState_Set(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		st.Set(i, i)
	}
}

func BenchmarkMapState_Get(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i, i*10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		_, _ = st.Get(i % 500)
	}
}

func BenchmarkMapState_Get_Miss(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i, i*10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		_, _ = st.Get(i + 1000) // Keys that don't exist
	}
}

func BenchmarkMapState_Has(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i, i*10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		_ = st.Has(i % 500)
	}
}

func BenchmarkMapState_Del(b *testing.B) {
	st := state.NewMapState[int, int](b.N)

	// Pre-populate with data to delete
	for i := range b.N {
		st.Set(i, i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		st.Del(i)
	}
}

func BenchmarkMapState_SetParallel(b *testing.B) {
	st := state.NewMapState[int, int](10000)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			st.Set(i, i)
			i++
		}
	})
}

func BenchmarkMapState_GetParallel(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i, i*10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = st.Get(i % 500)
			i++
		}
	})
}

func BenchmarkMapState_HasParallel(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate with some data
	for i := range 500 {
		st.Set(i, i*10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = st.Has(i % 500)
			i++
		}
	})
}

func BenchmarkMapState_MixedReadWrite(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate
	for i := range 100 {
		st.Set(i, i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				st.Set(i, i)
			} else {
				_, _ = st.Get(i % 100)
			}
			i++
		}
	})
}

func BenchmarkMapState_MixedAllOperations(b *testing.B) {
	st := state.NewMapState[int, int](1000)

	// Pre-populate
	for i := range 100 {
		st.Set(i, i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				st.Set(i, i)
			case 1:
				_, _ = st.Get(i % 100)
			case 2:
				_ = st.Has(i % 100)
			case 3:
				st.Del(i % 100)
			}
			i++
		}
	})
}

func BenchmarkMapState_Set_VaryingCapacity(b *testing.B) {
	capacities := []int{10, 100, 1000, 10000}

	for _, cap := range capacities {
		b.Run(fmt.Sprintf("capacity_%d", cap), func(b *testing.B) {
			st := state.NewMapState[int, int](cap)

			b.ResetTimer()
			b.ReportAllocs()
			for i := range b.N {
				st.Set(i, i)
			}
		})
	}
}

func BenchmarkMapState_StringKey(b *testing.B) {
	st := state.NewMapState[string, int](1000)
	keys := make([]string, 100)
	for i := range 100 {
		keys[i] = fmt.Sprintf("key_%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		st.Set(keys[i%100], i)
	}
}

func BenchmarkMapState_StringKey_Get(b *testing.B) {
	st := state.NewMapState[string, int](1000)
	keys := make([]string, 100)
	for i := range 100 {
		keys[i] = fmt.Sprintf("key_%d", i)
		st.Set(keys[i], i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		_, _ = st.Get(keys[i%100])
	}
}
