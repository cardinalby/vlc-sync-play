package cliargs

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	NotTagged int
	Str       string   `flag:"s" usage:"usage1"`
	Int       int      `flag:"i" usage:"usage2"`
	Bool      bool     `flag:"b" usage:"usage3"`
	StrP      *string  `flag:"sp" usage:"usage1"`
	IntP      *int     `flag:"ip" usage:"usage2"`
	BoolP     *bool    `flag:"bp" usage:"usage3"`
	Extra     []string `flagArgs:"true"`
}

func TestParseStructAllMandatory(t *testing.T) {
	s := testStruct{
		NotTagged: 1,
		Str:       "val1",
		Int:       3,
		Bool:      false,
	}
	err := ParseStruct(&s, "", flag.ExitOnError, []string{"--s", "val2", "--i", "2", "--b"})
	require.NoError(t, err)
	require.Equal(t, 1, s.NotTagged)
	require.Equal(t, "val2", s.Str)
	require.Equal(t, 2, s.Int)
	require.Equal(t, true, s.Bool)
	require.Nil(t, s.StrP)
	require.Nil(t, s.IntP)
	require.Nil(t, s.BoolP)
	require.Empty(t, s.Extra)
}

func TestParseStructSomeMandatory(t *testing.T) {
	s := testStruct{
		NotTagged: 1,
		Str:       "val1",
		Int:       2,
		Bool:      true,
	}
	err := ParseStruct(&s, "", flag.ExitOnError, []string{"--s", "val2", "abc", "def"})
	require.NoError(t, err)
	require.Equal(t, 1, s.NotTagged)
	require.Equal(t, "val2", s.Str)
	require.Equal(t, 2, s.Int)
	require.Equal(t, true, s.Bool)
	require.Nil(t, s.StrP)
	require.Nil(t, s.IntP)
	require.Nil(t, s.BoolP)
	require.Equal(t, []string{"abc", "def"}, s.Extra)
}

func TestParseStructAllOptional(t *testing.T) {
	s := testStruct{
		NotTagged: 1,
	}
	err := ParseStruct(&s, "", flag.ExitOnError, []string{"--sp", "val2", "--ip", "2", "--bp"})
	require.NoError(t, err)
	require.Equal(t, 1, s.NotTagged)
	require.NotNil(t, s.StrP)
	require.Equal(t, "val2", *s.StrP)
	require.NotNil(t, s.IntP)
	require.Equal(t, 2, *s.IntP)
	require.NotNil(t, s.BoolP)
	require.Equal(t, true, *s.BoolP)
}

func TestParseStructSomeOptional(t *testing.T) {
	s := testStruct{
		NotTagged: 1,
	}
	err := ParseStruct(&s, "", flag.ExitOnError, []string{"--sp", "val2"})
	require.NoError(t, err)
	require.Equal(t, 1, s.NotTagged)
	require.NotNil(t, s.StrP)
	require.Equal(t, "val2", *s.StrP)
	require.Nil(t, s.IntP)
	require.Nil(t, s.BoolP)
}
