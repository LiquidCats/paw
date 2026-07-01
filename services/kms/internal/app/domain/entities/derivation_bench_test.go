package entities_test

import (
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
)

func BenchmarkNewIndex(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = entities.NewIndex(44, true)
	}
}

func BenchmarkIndexFromUint32_Normal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = entities.IndexFromUint32(42)
	}
}

func BenchmarkIndexFromUint32_Hardened(b *testing.B) {
	b.ReportAllocs()
	raw := uint32(44) + entities.HardenedKeyIndex
	for b.Loop() {
		_ = entities.IndexFromUint32(raw)
	}
}

func BenchmarkIndex_Uint32(b *testing.B) {
	idx := entities.NewIndex(44, true)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = idx.Uint32()
	}
}

func BenchmarkIndex_IsHardened(b *testing.B) {
	idx := entities.NewIndex(44, true)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = idx.IsHardened()
	}
}

func BenchmarkIndex_Incr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		idx := entities.NewIndex(0, false)
		idx.Incr()
	}
}

func BenchmarkIndex_String_Normal(b *testing.B) {
	idx := entities.NewIndex(44, false)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = idx.String()
	}
}

func BenchmarkIndex_String_Hardened(b *testing.B) {
	idx := entities.NewIndex(44, true)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = idx.String()
	}
}

func BenchmarkIndex_UnmarshalText_Normal(b *testing.B) {
	input := []byte("44")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalText(input)
	}
}

func BenchmarkIndex_UnmarshalText_Hardened(b *testing.B) {
	input := []byte("44'")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalText(input)
	}
}

func BenchmarkIndex_UnmarshalJSON_Normal(b *testing.B) {
	input := []byte(`"44"`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalJSON(input)
	}
}

func BenchmarkIndex_UnmarshalJSON_Hardened(b *testing.B) {
	input := []byte(`"44'"`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalJSON(input)
	}
}

func BenchmarkIndex_UnmarshalYAML_Normal(b *testing.B) {
	input := []byte(`"44"`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalYAML(input)
	}
}

func BenchmarkIndex_UnmarshalYAML_Hardened(b *testing.B) {
	input := []byte(`"44'"`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalYAML(input)
	}
}

func BenchmarkIndex_RoundTrip_StringUnmarshalText(b *testing.B) {
	original := entities.NewIndex(44, true)
	str := []byte(original.String())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var idx entities.Index
		_ = idx.UnmarshalText(str)
	}
}
