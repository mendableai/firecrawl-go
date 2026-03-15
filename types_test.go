package firecrawl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOrStringSlice_SingleString(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`"hello"`))
	require.NoError(t, err)
	assert.Equal(t, StringOrStringSlice{"hello"}, s)
}

func TestStringOrStringSlice_StringArray(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`["a","b","c"]`))
	require.NoError(t, err)
	assert.Equal(t, StringOrStringSlice{"a", "b", "c"}, s)
}

func TestStringOrStringSlice_EmptyArray(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`[]`))
	require.NoError(t, err)
	assert.Equal(t, StringOrStringSlice{}, s)
}

func TestStringOrStringSlice_EmptyString(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`""`))
	require.NoError(t, err)
	assert.Equal(t, StringOrStringSlice{""}, s)
}

func TestStringOrStringSlice_InvalidType_Number(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`123`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "neither a string nor a list of strings")
}

func TestStringOrStringSlice_InvalidType_Boolean(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`true`))
	assert.Error(t, err)
}

func TestStringOrStringSlice_InvalidType_Object(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`{"key":"value"}`))
	assert.Error(t, err)
}

func TestStringOrStringSlice_Null(t *testing.T) {
	var s StringOrStringSlice
	err := s.UnmarshalJSON([]byte(`null`))
	// JSON null unmarshals into a string as "" (zero value) — so the first branch succeeds.
	// The result is a slice containing an empty string.
	require.NoError(t, err)
	assert.Equal(t, StringOrStringSlice{""}, s)
}
