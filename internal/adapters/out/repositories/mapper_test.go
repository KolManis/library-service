package repositories

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapSlice(t *testing.T) {
	ints := []int{1, 2, 3}

	strs := mapSlice(ints, func(i int) string {
		return string(rune('a' + i - 1))
	})

	require.Equal(t, []string{"a", "b", "c"}, strs)
}

func TestMapSlice_Empty(t *testing.T) {
	var empty []int

	result := mapSlice(empty, func(i int) int { return i })

	require.Empty(t, result)
}
