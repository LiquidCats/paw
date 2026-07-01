package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/stretchr/testify/suite"
)

// ── Index Suite ───────────────────────────────────────────────────────────────

type IndexSuite struct{ suite.Suite }

func TestIndexSuite(t *testing.T) { suite.Run(t, new(IndexSuite)) }

// ── NewIndex ──────────────────────────────────────────────────────────────────

func (s *IndexSuite) TestNewIndex_Normal() {
	idx := entities.NewIndex(5, false)
	s.Equal(uint32(5), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestNewIndex_Hardened() {
	idx := entities.NewIndex(5, true)
	s.Equal(uint32(5)+entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestNewIndex_Zero() {
	idx := entities.NewIndex(0, false)
	s.Equal(uint32(0), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestNewIndex_HardenedZero() {
	idx := entities.NewIndex(0, true)
	s.Equal(entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

// ── IndexFromUint32 ───────────────────────────────────────────────────────────

func (s *IndexSuite) TestIndexFromUint32_Normal() {
	idx := entities.IndexFromUint32(42)
	s.Equal(uint32(42), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestIndexFromUint32_Hardened() {
	raw := uint32(0) + entities.HardenedKeyIndex
	idx := entities.IndexFromUint32(raw)
	s.Equal(raw, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestIndexFromUint32_HardenedWithOffset() {
	raw := uint32(44) + entities.HardenedKeyIndex
	idx := entities.IndexFromUint32(raw)
	s.Equal(raw, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestIndexFromUint32_ExactBoundary() {
	// HardenedKeyIndex itself should produce hardened index 0
	idx := entities.IndexFromUint32(entities.HardenedKeyIndex)
	s.True(idx.IsHardened())
	s.Equal(entities.HardenedKeyIndex, idx.Uint32())
}

func (s *IndexSuite) TestIndexFromUint32_BelowBoundary() {
	idx := entities.IndexFromUint32(entities.HardenedKeyIndex - 1)
	s.False(idx.IsHardened())
	s.Equal(entities.HardenedKeyIndex-1, idx.Uint32())
}

// ── Uint32 ────────────────────────────────────────────────────────────────────

func (s *IndexSuite) TestUint32_RoundTrip_Normal() {
	idx := entities.NewIndex(100, false)
	s.Equal(uint32(100), idx.Uint32())
}

func (s *IndexSuite) TestUint32_RoundTrip_Hardened() {
	idx := entities.NewIndex(100, true)
	s.Equal(uint32(100)+entities.HardenedKeyIndex, idx.Uint32())
}

// ── IsHardened ────────────────────────────────────────────────────────────────

func (s *IndexSuite) TestIsHardened_False() {
	s.False(entities.NewIndex(1, false).IsHardened())
}

func (s *IndexSuite) TestIsHardened_True() {
	s.True(entities.NewIndex(1, true).IsHardened())
}

// ── Incr ──────────────────────────────────────────────────────────────────────

func (s *IndexSuite) TestIncr_Normal() {
	idx := entities.NewIndex(3, false)
	idx.Incr()
	s.Equal(uint32(4), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestIncr_Hardened_PreservesHardened() {
	idx := entities.NewIndex(0, true)
	idx.Incr()
	s.Equal(uint32(1)+entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestIncr_Multiple() {
	idx := entities.NewIndex(0, false)
	for range 5 {
		idx.Incr()
	}
	s.Equal(uint32(5), idx.Uint32())
}

// ── String ────────────────────────────────────────────────────────────────────

func (s *IndexSuite) TestString_Normal() {
	s.Equal("0", entities.NewIndex(0, false).String())
	s.Equal("44", entities.NewIndex(44, false).String())
}

func (s *IndexSuite) TestString_Hardened() {
	s.Equal("0'", entities.NewIndex(0, true).String())
	s.Equal("44'", entities.NewIndex(44, true).String())
}

// ── UnmarshalText ─────────────────────────────────────────────────────────────

func (s *IndexSuite) TestUnmarshalText_Normal() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalText([]byte("7")))
	s.Equal(uint32(7), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalText_Hardened() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalText([]byte("44'")))
	s.Equal(uint32(44)+entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalText_Zero() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalText([]byte("0")))
	s.Equal(uint32(0), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalText_Empty_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte(""))
	s.Require().Error(err)
	s.ErrorContains(err, "cannot unmarshal an empty Index")
}

func (s *IndexSuite) TestUnmarshalText_Null_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("null"))
	s.Require().Error(err)
}

func (s *IndexSuite) TestUnmarshalText_Nil_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("nil"))
	s.Require().Error(err)
}

func (s *IndexSuite) TestUnmarshalText_EmptyObject_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("{}"))
	s.Require().Error(err)
}

func (s *IndexSuite) TestUnmarshalText_EmptyArray_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("[]"))
	s.Require().Error(err)
}

func (s *IndexSuite) TestUnmarshalText_NonNumeric_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("abc"))
	s.Require().Error(err)
	s.ErrorContains(err, "cannot unmarshal an Index")
}

func (s *IndexSuite) TestUnmarshalText_NegativeNumber_Errors() {
	var idx entities.Index
	err := idx.UnmarshalText([]byte("-1"))
	s.Require().Error(err)
}

// ── UnmarshalJSON ─────────────────────────────────────────────────────────────

func (s *IndexSuite) TestUnmarshalJSON_Normal() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalJSON([]byte(`"12"`)))
	s.Equal(uint32(12), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalJSON_Hardened() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalJSON([]byte(`"44'"`)))
	s.Equal(uint32(44)+entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalJSON_EmptyString_Errors() {
	var idx entities.Index
	err := idx.UnmarshalJSON([]byte(`""`))
	s.Require().Error(err)
}

func (s *IndexSuite) TestUnmarshalJSON_InvalidJSON_Errors() {
	var idx entities.Index
	err := idx.UnmarshalJSON([]byte(`not-json`))
	s.Require().Error(err)
	s.ErrorContains(err, "cannot unmarshal Index from json")
}

func (s *IndexSuite) TestUnmarshalJSON_ViaStdLib() {
	type wrapper struct {
		Idx entities.Index `json:"idx"`
	}
	var w wrapper
	s.Require().NoError(json.Unmarshal([]byte(`{"idx":"0'"}`), &w))
	s.True(w.Idx.IsHardened())
	s.Equal(entities.HardenedKeyIndex, w.Idx.Uint32())
}

// ── UnmarshalYAML ─────────────────────────────────────────────────────────────

func (s *IndexSuite) TestUnmarshalYAML_Normal() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalYAML([]byte(`"5"`)))
	s.Equal(uint32(5), idx.Uint32())
	s.False(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalYAML_Hardened() {
	var idx entities.Index
	s.Require().NoError(idx.UnmarshalYAML([]byte(`"60'"`)))
	s.Equal(uint32(60)+entities.HardenedKeyIndex, idx.Uint32())
	s.True(idx.IsHardened())
}

func (s *IndexSuite) TestUnmarshalYAML_EmptyString_Errors() {
	var idx entities.Index
	err := idx.UnmarshalYAML([]byte(`""`))
	s.Require().Error(err)
}

// ── String ↔ UnmarshalText round-trip ────────────────────────────────────────

func (s *IndexSuite) TestRoundTrip_StringUnmarshalText_Normal() {
	original := entities.NewIndex(99, false)
	var recovered entities.Index
	s.Require().NoError(recovered.UnmarshalText([]byte(original.String())))
	s.Equal(original.Uint32(), recovered.Uint32())
	s.Equal(original.IsHardened(), recovered.IsHardened())
}

func (s *IndexSuite) TestRoundTrip_StringUnmarshalText_Hardened() {
	original := entities.NewIndex(44, true)
	var recovered entities.Index
	s.Require().NoError(recovered.UnmarshalText([]byte(original.String())))
	s.Equal(original.Uint32(), recovered.Uint32())
	s.Equal(original.IsHardened(), recovered.IsHardened())
}

func (s *IndexSuite) TestRoundTrip_IndexFromUint32_Uint32() {
	for _, raw := range []uint32{0, 1, 100, entities.HardenedKeyIndex, entities.HardenedKeyIndex + 44} {
		idx := entities.IndexFromUint32(raw)
		s.Equal(raw, idx.Uint32(), "raw=%d", raw)
	}
}
